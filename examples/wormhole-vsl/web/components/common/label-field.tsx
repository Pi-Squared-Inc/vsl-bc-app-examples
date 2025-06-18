import clsx from "clsx";
import React, { ReactNode } from "react";

interface LabelFieldProps {
  label: ReactNode;
  value: ReactNode;
  valueClassName?: string;
  valueType?: "code" | "normal";
}

const LabelField: React.FC<LabelFieldProps> = ({
  label,
  value,
  valueClassName,
  valueType = "code",
}) => {
  return (
    <div className="flex flex-col gap-2">
      <div className="text-sm font-semibold">{label}</div>
      {valueType === "code" ? (
        <pre
          className={clsx(
            "bg-card border p-2 rounded-md text-base whitespace-break-spaces break-all max-h-[200px] overflow-y-auto",
            valueClassName
          )}
        >
          {value}
        </pre>
      ) : (
        <div
          className={clsx(
            "bg-card border p-2 rounded-md text-base whitespace-break-spaces break-all max-h-[200px] overflow-y-auto",
            valueClassName
          )}
        >
          {value}
        </div>
      )}
    </div>
  );
};

export default LabelField;
