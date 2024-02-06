import {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { FC, PropsWithChildren } from "react";

import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

type ModalWrapperProps = PropsWithChildren & {
  title: string;
  formId?: string;
  showSaveBtn?: boolean;
  onCancel: () => void;
};

export const ModalWrapper: FC<ModalWrapperProps> = ({
  title,
  showSaveBtn = true,
  children,
  formId,
  onCancel,
}) => {
  const { t } = useTranslation();
  return (
    <DialogContent className="sm:max-w-xl">
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
      </DialogHeader>
      <div className="flex max-h-[70vh] flex-col gap-5 overflow-y-auto p-[1px]">
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
