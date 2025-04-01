import {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { FC, PropsWithChildren } from "react";

import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type ModalWrapperProps = PropsWithChildren & {
  title: string;
  formId?: string;
  showSaveBtn?: boolean;
  onCancel: () => void;
  size?: "md" | "lg";
};

export const ModalWrapper: FC<ModalWrapperProps> = ({
  title,
  showSaveBtn = true,
  children,
  formId,
  onCancel,
  size = "md",
}) => {
  const { t } = useTranslation();

  return (
    <DialogContent
      className={twMergeClsx(
        size === "md" && "sm:max-w-xl",
        size === "lg" && "sm:max-w-4xl"
      )}
    >
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
      </DialogHeader>
      <div className="flex max-h-[70vh] flex-col gap-5 overflow-y-auto p-px">
        {children}
      </div>
      <DialogFooter>
        <Button type="button" variant="ghost" onClick={onCancel}>
          {t("components.modalWrapper.cancelBtn")}
        </Button>
        {showSaveBtn && (
          <Button type="submit" form={formId ?? undefined}>
            {t("components.modalWrapper.saveBtn")}
          </Button>
        )}
      </DialogFooter>
    </DialogContent>
  );
};
