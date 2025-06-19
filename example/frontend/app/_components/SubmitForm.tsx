"use client";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { VerificationRecord } from "@/types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAtom } from "jotai";
import Image from "next/image";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { toHex } from "viem";
import { useAccount, useSignMessage } from "wagmi";
import { z } from "zod";
import { generateSignatureComponents, hashFileSha256, hashStringSha256 } from "../../utils/signature";
import ExampleImage1 from "../_assets/goldfish.jpeg";
import { getBackendAddress, checkCanSubmit } from "../actions/validationRecord";
import { getAccountNonce, pay } from "../actions/vsl";
import { fetchBalanceAtom } from "../store/balance";

const fileSizeLimit = 10 * 1024 * 1024; // 10MB
const textCharLimit = 100;

const formSchema = z
  .object({
    type: z.union([z.literal("img_class"), z.literal("plain_text")]),
    image: z.instanceof(File).optional(),
    text: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.type === "img_class") {
      if (!data.image) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Image is required for image classification",
          path: ["image"],
        });
      } else {
        if (
          !["image/png", "image/jpeg", "image/jpg"].includes(data.image.type)
        ) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Invalid image file type",
            path: ["image"],
          });
        }
        if (data.image.size > fileSizeLimit) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "File size should not exceed 5MB",
            path: ["image"],
          });
        }
      }
    }
    if (data.type === "plain_text") {
      if (!data.text) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Text is required for plain text processing",
          path: ["text"],
        });
      } else if (data.text.length > textCharLimit) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message:
            "Prompt exceeds maximum length (" +
            data.text.length.toString() +
            " / " +
            textCharLimit.toString() +
            ")",
          path: ["text"],
        });
      }
    }
  });

type FormValues = z.infer<typeof formSchema>;

const defaultValues: FormValues = {
  type: "img_class",
  image: undefined,
  text: "",
};

export default function SubmitForm({
  refetch,
  className,
}: {
  refetch?: (backToFirst?: boolean) => Promise<void>;
  className?: string;
}) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    mode: "onChange",
    reValidateMode: "onChange",
    defaultValues,
  });
  const [isComputing, setIsComputing] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  const { isConnected, address } = useAccount();
  const [, fetchBalance] = useAtom(fetchBalanceAtom);
  const { signMessageAsync } = useSignMessage();

  const handleFileChange = (
    file: File | undefined,
    fieldOnChange: (file?: File) => void
  ) => {
    if (!file) {
      setPreviewUrl(null);
      fieldOnChange(undefined);
      return;
    }
    if (
      !["image/png", "image/jpeg", "image/jpg"].includes(file.type) ||
      file.size > fileSizeLimit
    ) {
      setPreviewUrl(null);
      fieldOnChange(undefined);
      return;
    }
    setPreviewUrl(URL.createObjectURL(file));
    fieldOnChange(file);
  };

  const onSubmit = async (values: FormValues) => {
    try {
      setIsComputing(true);

      const canSubmit = (await checkCanSubmit(address!, values.type)) as {
        allowed: boolean;
      };
      if (!canSubmit.allowed) {
        throw new Error("please wait for your other " + values.type + " request to finish");
      }      

      const backendAddressResponse = (await getBackendAddress()) as {
        address: string;
      };
      const backendAddress = backendAddressResponse.address;
      const nonce = await getAccountNonce(address!);
      const amount = "0x" + BigInt(20 * (10**18)).toString(16);

      const paySignatureComponents = await generateSignatureComponents(
        async ({ message }) => {
          return await signMessageAsync({ message });
        },
        [
          toHex(address!),
          toHex(backendAddress),
          toHex(amount),
          toHex(nonce.toString()),
        ]
      );

      const paymentClaimId = await pay(
        address!,
        backendAddress,
        amount,
        nonce.toString(),
        paySignatureComponents
      );

      const formData = new FormData();
      let hashedInput : string = "";
      let computeType : string = "";
      formData.append("type", values.type);
      formData.append("sender_address", address!);
      formData.append("payment_claim_id", paymentClaimId);
      if (values.type === "img_class" && values.image) {
        formData.append("image", values.image);
        computeType = "img_class";
        hashedInput = await hashFileSha256(values.image!);
      }
      if (values.type === "plain_text" && values.text) {
        formData.append("prompt", values.text);
        computeType = "text_gen";
        hashedInput = await hashStringSha256(values.text);
      }
      const computeSignatureComponents = await generateSignatureComponents(
        async ({ message }) => {
          return await signMessageAsync({ message });
        },
        [
          toHex(computeType),
          toHex(address!),
          toHex(paymentClaimId),
          toHex(hashedInput)
        ]
      );
      formData.append("hash", computeSignatureComponents.hash)
      formData.append("r", computeSignatureComponents.r)
      formData.append("s", computeSignatureComponents.s)
      formData.append("v", computeSignatureComponents.v.toString())
      
      const response = await fetch(
        process.env.NEXT_PUBLIC_API_URL + "/compute",
        {
          method: "POST",
          body: formData,
        }
      );

      if (!response.ok) {
        const backend_error = await response.text()
        throw new Error(backend_error);
      }

      const result = (await response.json()) as VerificationRecord;

      if (result.id && refetch) {
        form.setValue("image", undefined);
        form.setValue("text", "");
        setPreviewUrl(null);
        await fetchBalance(address!);
        await refetch(true);
      }
    } catch (error) {
      toast.error("Error during computation: " + error);
    } finally {
      setIsComputing(false);
    }
  };

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("space-y-8", className)}
      >
        <FormField
          control={form.control}
          name="type"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Type</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a type" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="img_class">
                    Image Classification
                  </SelectItem>
                  <SelectItem value="plain_text">Plain Text</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />
        {form.watch("type") === "img_class" && (
          <FormField
            control={form.control}
            name="image"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Image</FormLabel>
                <div>
                  <label className="relative flex flex-col items-center justify-center w-full min-h-24 border-2 border-dashed rounded cursor-pointer hover:border-primary transition-colors p-4">
                    {previewUrl ? (
                      <Image
                        src={previewUrl}
                        alt={field.value?.name || "Selected image"}
                        className="w-20 h-20 object-cover rounded border mb-1"
                        width={80}
                        height={80}
                        onError={() => setPreviewUrl(null)}
                      />
                    ) : (
                      <span className="text-sm text-muted-foreground">
                        Click or drag an image here to upload
                      </span>
                    )}
                    <span className="text-xs text-muted-foreground mt-2">
                      {field.value?.name
                        ? `Selected file: ${field.value.name}`
                        : ""}
                    </span>
                    <input
                      type="file"
                      accept="image/png, image/jpeg, image/jpg"
                      className="absolute inset-0 opacity-0 cursor-pointer"
                      style={{ width: "100%", height: "100%" }}
                      onChange={(e) => {
                        const files = e.target.files;
                        if (files && files[0]) {
                          handleFileChange(files[0], field.onChange);
                        } else {
                          setPreviewUrl(null);
                          field.onChange(undefined);
                        }
                      }}
                      onDrop={(e) => {
                        const files = e.dataTransfer.files;
                        if (files && files[0]) {
                          handleFileChange(files[0], field.onChange);
                        }
                      }}
                    />
                  </label>
                </div>
                <div className="flex gap-2 mt-2">
                  {!previewUrl && (
                    <button
                      type="button"
                      onClick={async () => {
                        const response = await fetch(ExampleImage1.src);
                        const blob = await response.blob();
                        const file = new File([blob], "goldfish.jpeg", {
                          type: blob.type,
                        });
                        setPreviewUrl(URL.createObjectURL(file));
                        field.onChange(file);
                      }}
                      tabIndex={0}
                      aria-label="Use example image"
                    >
                      <Image
                        src={ExampleImage1.src}
                        alt="goldfish"
                        className="w-20 h-20 object-cover rounded border cursor-pointer"
                        width={ExampleImage1.width}
                        height={ExampleImage1.height}
                      />
                    </button>
                  )}
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
        {form.watch("type") === "plain_text" && (
          <FormField
            control={form.control}
            name="text"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Text</FormLabel>
                <Textarea
                  placeholder="Enter text for processing"
                  value={field.value || ""}
                  onChange={(e) => field.onChange(e.target.value)}
                />
                <FormMessage />
              </FormItem>
            )}
          />
        )}
        <div className="flex justify-end flex-col items-end gap-2">
          <Button
            type="submit"
            className="w-full"
            disabled={!form.formState.isValid || isComputing || !isConnected}
          >
            {isComputing ? "Computing..." : "Confirm"}
          </Button>
        </div>
      </form>
    </Form>
  );
}
