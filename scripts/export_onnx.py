import argparse
import os

import onnx
import torch


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--ckpt", default="checkpoints/model.pt")
    parser.add_argument("--out", required=True)
    parser.add_argument("--opset", type=int, default=17)
    args = parser.parse_args()

    model = torch.nn.Linear(34 * 32, 128)

    if not os.path.exists(args.ckpt):
        ckpt_dir = os.path.dirname(args.ckpt)
        if ckpt_dir:
            os.makedirs(ckpt_dir, exist_ok=True)
        torch.save(model.state_dict(), args.ckpt)
    else:
        state = torch.load(args.ckpt, map_location="cpu")
        model.load_state_dict(state)
    model.eval()
    dummy = torch.randn(1, 34 * 32)

    torch.onnx.export(
        model,
        dummy,
        args.out,
        opset_version=args.opset,
        input_names=["input"],
        output_names=["output"],
    )

    loaded = onnx.load(args.out)
    onnx.checker.check_model(loaded)
    print("ONNX saved:", args.out)


if __name__ == "__main__":
    main()
