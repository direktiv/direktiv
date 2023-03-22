import * as React from "react";

import clsx from "clsx";

// const buttonVariants = clsx(
//   "active:scale-95 inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2 dark:hover:bg-slate-800 dark:hover:text-slate-100 disabled:opacity-50 dark:focus:ring-slate-400 disabled:pointer-events-none dark:focus:ring-offset-slate-900 data-[state=open]:bg-slate-100 dark:data-[state=open]:bg-slate-800",
//   {
//     variants: {
//       variant: {
//         default:
//           "bg-slate-900 text-white hover:bg-slate-700 dark:bg-slate-50 dark:text-slate-900",
//         destructive:
//           "bg-red-500 text-white hover:bg-red-600 dark:hover:bg-red-600",
//         outline:
//           "bg-transparent border border-slate-200 hover:bg-slate-100 dark:border-slate-700 dark:text-slate-100",
//         subtle:
//           "bg-slate-100 text-slate-900 hover:bg-slate-200 dark:bg-slate-700 dark:text-slate-100",
//         ghost:
//           "bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 dark:text-slate-100 dark:hover:text-slate-100 data-[state=open]:bg-transparent dark:data-[state=open]:bg-transparent",
//         link: "bg-transparent dark:bg-transparent underline-offset-4 hover:underline text-slate-900 dark:text-slate-100 hover:bg-transparent dark:hover:bg-transparent",
//       },
//       size: {
//         default: "h-10 py-2 px-4",
//         sm: "h-9 px-2 rounded-md",
//         lg: "h-11 px-8 rounded-md",
//       },
//     },
//     defaultVariants: {
//       variant: "default",
//       size: "default",
//     },
//   }
// );

type ButtonProps = {
  variant?: "destructive" | "outline" | "primary" | "ghost" | "link";
  size?: "xs" | "sm" | "lg";
  loading?: boolean;
  circle?: boolean;
};

const Button = React.forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement> & ButtonProps
>(({ className, variant, size, ...props }, ref) => (
  <button
    className={clsx(
      "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2 active:scale-95 disabled:pointer-events-none     disabled:opacity-50 data-[state=open]:bg-slate-100 dark:hover:bg-slate-800 dark:hover:text-slate-100 dark:focus:ring-slate-400 dark:focus:ring-offset-slate-900 dark:data-[state=open]:bg-slate-800",
      !variant &&
        "bg-slate-900 text-white hover:bg-slate-700 dark:bg-slate-50 dark:text-slate-900",
      variant === "destructive" &&
        "bg-red-500 text-white hover:bg-red-600 dark:hover:bg-red-600",
      variant === "outline" &&
        "border border-slate-200 bg-transparent hover:bg-slate-100 dark:border-slate-700 dark:text-slate-100",
      variant === "primary" &&
        "bg-slate-100 text-slate-900 hover:bg-slate-200 dark:bg-slate-700 dark:text-slate-100",
      variant === "ghost" &&
        "bg-transparent hover:bg-slate-100 data-[state=open]:bg-transparent dark:text-slate-100 dark:hover:bg-slate-800 dark:hover:text-slate-100 dark:data-[state=open]:bg-transparent",
      variant === "link" &&
        "bg-transparent text-slate-900 underline-offset-4 hover:bg-transparent hover:underline dark:bg-transparent dark:text-slate-100 dark:hover:bg-transparent",
      size === "xs" && "h-8 rounded-md px-2",
      size === "sm" && "h-9 rounded-md px-2",
      !size && "h-10 py-2 px-4",
      size === "lg" && "h-11 rounded-md px-8",
      className
    )}
    ref={ref}
    {...props}
  />
));
Button.displayName = "Button";

export default Button;
