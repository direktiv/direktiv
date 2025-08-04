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
    <fieldset
      className={twMergeClsx(
        "mb-2 flex gap-1",
        horizontal
          ? "flex-row-reverse flex-wrap items-center gap-2"
          : "flex-col gap-1"
      )}
    >
      <label className="flex grow text-sm font-bold" htmlFor={htmlFor}>
        <span className="grow">{label}</span>
        {optional && (
          <span className="font-normal text-gray-9 dark:text-gray-dark-9">
            {" "}
            {t("direktivPage.page.blocks.form.optional")}
          </span>
        )}
      </label>
      {children}
      {description && (
        <div className="basis-full text-sm text-gray-9 dark:text-gray-dark-9">
          {description}
        </div>
      )}
    </fieldset>
  );
};
