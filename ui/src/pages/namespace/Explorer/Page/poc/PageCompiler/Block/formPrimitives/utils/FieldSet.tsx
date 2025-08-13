import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type FieldsetProps = PropsWithChildren & {
  label: string;
  description: string;
  htmlFor: string;
  optional: boolean;
  horizontal?: boolean;
};

export const Fieldset: FC<FieldsetProps> = ({
  label,
  description,
  children,
  htmlFor,
  horizontal,
  optional,
}) => {
  const { t } = useTranslation();
  return (
    <fieldset className="flex flex-col gap-1">
      <label className="flex grow gap-1 text-sm font-bold" htmlFor={htmlFor}>
        <span>{label}</span>
        {optional && (
          <span className="font-normal text-gray-9 dark:text-gray-dark-9">
            {" "}
            {t("direktivPage.page.blocks.form.optional")}
          </span>
        )}
      </label>
      <div
        className={twMergeClsx(
          "flex min-h-9",
          horizontal ? "flex-row items-center gap-3" : "flex-col gap-1"
        )}
      >
        {children}
        {description && (
          <div className="basis-full text-sm text-gray-9 dark:text-gray-dark-9">
            {description}
          </div>
        )}
      </div>
    </fieldset>
  );
};
