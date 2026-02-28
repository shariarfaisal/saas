import { SelectHTMLAttributes, forwardRef } from "react";

import { cn } from "@/lib/utils";

export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className, children, ...props }, ref) => {
    return (
      <select
        className={cn("w-full rounded-md border border-slate-300 px-3 py-2 text-sm", className)}
        ref={ref}
        {...props}
      >
        {children}
      </select>
    );
  },
);

Select.displayName = "Select";
