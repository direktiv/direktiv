import { FC, MouseEvent, PropsWithChildren } from "react";

import { TemplateString } from "../../../primitives/TemplateString";
import { twMergeClsx } from "~/util/helpers";
import { useFormValidationContext } from "../../Form/FormValidationContext";
import { useTranslation } from "react-i18next";

type FieldsetProps = PropsWithChildren & {
  id: string;
  label: string;
  description: string;
  htmlFor: string;
  optional: boolean;
  horizontal?: boolean;
  onClickLabel?: (event: MouseEvent<HTMLElement>) => void;
};

export const Fieldset: FC<FieldsetProps> = ({
  id,
  label,
  description,
  children,
  htmlFor,
  horizontal,
  optional,
  onClickLabel,
}) => {
  const { t } = useTranslation();
  const { missingFields } = useFormValidationContext();
  const isMissingField = missingFields.includes(id);

  return (
    <fieldset
      className={twMergeClsx(
        "flex flex-col gap-1",
        isMissingField &&
          "rounded-sm outline outline-2 outline-offset-8 outline-danger-7 dark:outline-danger-dark-7"
      )}
    >
      <label
        className={twMergeClsx(
          "flex grow gap-1 text-sm font-bold",
          isMissingField && "text-danger-11"
        )}
        htmlFor={htmlFor}
        onClick={onClickLabel}
      >
        <span>
          <TemplateString value={label} />
        </span>
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
            <TemplateString value={description} />
          </div>
        )}
      </div>
    </fieldset>
  );
};
