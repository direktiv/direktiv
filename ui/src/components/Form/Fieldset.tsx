import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

type FieldsetProps = PropsWithChildren & {
  label: string;
  htmlFor?: string;
  className?: string;
  horizontal?: boolean;
};

export const Fieldset: FC<FieldsetProps> = ({
  label,
  htmlFor,
  children,
  className,
  horizontal,
}) => (
  <fieldset
    className={twMergeClsx(
      "mb-2 flex gap-2",
      className,
      horizontal ? "flex-row-reverse items-center" : "flex-col"
    )}
  >
    <label className="grow text-sm" htmlFor={htmlFor}>
      {label}
    </label>
    {children}
  </fieldset>
);
