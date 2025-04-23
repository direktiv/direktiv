import { ElementRef, forwardRef } from "react";
import { Loader2, LucideIcon, LucideProps } from "lucide-react";

import { twMergeClsx } from "~/util/helpers";

export const Loading = forwardRef<ElementRef<LucideIcon>, LucideProps>(
  ({ className, ...props }, ref) => (
    <Loader2
      className={twMergeClsx("size-4 animate-spin", className)}
      ref={ref}
      {...props}
    />
  )
);

Loading.displayName = "Loading";
