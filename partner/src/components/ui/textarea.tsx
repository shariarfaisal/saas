import { TextareaHTMLAttributes, forwardRef } from "react";

import { cn } from "@/lib/utils";

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        className={cn("w-full rounded-md border border-slate-300 px-3 py-2 text-sm", className)}
        ref={ref}
        {...props}
      />
    );
  },
);

Textarea.displayName = "Textarea";
