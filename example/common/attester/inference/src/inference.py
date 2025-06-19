import sys
import torch
from torchvision import models, transforms
from PIL import Image
import urllib.request
from ctransformers import AutoModelForCausalLM


def handle_resnet18(model_path, image_path):
    # 1. Define the model architecture
    # change the following in the cloud runtime
    model = models.resnet18()  # must match the original architecture
    model.eval()
    model.load_state_dict(torch.load(model_path, map_location="cpu"))

    # Load and preprocess the image
    image = Image.open(image_path).convert("RGB")
    transform = transforms.Compose([
        transforms.Resize((224, 224)),
        transforms.ToTensor()
    ])
    input_tensor = transform(image).unsqueeze(0)

    # Inference
    with torch.no_grad():
        output_tensor = model(input_tensor)
    
    # Download ImageNet labels
    url = "https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt"
    labels = urllib.request.urlopen(url).read().decode("utf-8").splitlines()

    # Get top-1 prediction
    _, predicted = output_tensor[0].topk(1)
    
    # Consolidate the output
    result = {'model_output': output_tensor, 'predicted_class': labels[predicted]}
    print(result)

def handle_llama(model_path, input_path):
    model = AutoModelForCausalLM.from_pretrained(model_path, model_type="llama")
    # Load very first pretrained model
    with open(input_path, "r") as fin:
        input_text = fin.read().strip()
    output_text = model(input_text, max_new_tokens=25, stop=["."]).lstrip()
    print(output_text)
    return output_text

def main(model_path, input_path):
    if model_path.endswith('.pt'):
        result = handle_resnet18(model_path, input_path)
    elif model_path.endswith('.gguf'):
        result = handle_llama(model_path, input_path)
    else:
        print('Expecting either .pt or .gguf model!')
        sys.exit(1)
    return result

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python3 inference.py <model_path> <input_file>")
        sys.exit(1)

    model_path = sys.argv[1]
    input_path = sys.argv[2]
    main(model_path, input_path)
