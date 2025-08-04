import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

type FieldsetProps = PropsWithChildren & {
  label: string;
  description: string;
  htmlFor: string;
  horizontal?: boolean;
};

export const Fieldset: FC<FieldsetProps> = ({
  label,
  description,
  children,
  htmlFor,
  horizontal,
}) => (
  <fieldset
    className={twMergeClsx(
      "mb-2 flex gap-2",
      horizontal ? "flex-row-reverse items-center" : "flex-col"
    )}
  >
    <label className="grow text-sm" htmlFor={htmlFor}>
      {label}
    </label>
    {children}
    <div>{description}</div>
  </fieldset>
);
