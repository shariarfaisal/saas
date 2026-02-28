import { InputHTMLAttributes, forwardRef } from "react";

import { cn } from "@/lib/utils";

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => {
    return (
      <input
        className={cn("w-full rounded-md border border-slate-300 px-3 py-2 text-sm", className)}
        ref={ref}
        {...props}
      />
    );
  },
);

Input.displayName = "Input";
