import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

type FieldsetProps = PropsWithChildren & {
  label: string;
  description: string;
  htmlFor: string;
  required: boolean;
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
      "mb-2 flex gap-1",
      horizontal
        ? "flex-row-reverse flex-wrap items-center gap-2"
        : "flex-col gap-1"
    )}
  >
    <label className="grow text-sm font-bold" htmlFor={htmlFor}>
      {label}
    </label>
    {children}
    {description && (
      <div className="basis-full text-sm text-gray-9 dark:text-gray-dark-9">
        {description}
      </div>
    )}
  </fieldset>
);
