import clsx, { ClassValue } from "clsx";

import { twMerge } from "tailwind-merge";

export const twMergeClsx = (...inputs: ClassValue[]) =>
  twMerge(clsx(...inputs));
