import {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { FC, PropsWithChildren } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { useTranslation } from "react-i18next";

type ModalWrapperProps = PropsWithChildren & {
  title: string;
  formId?: string;
  showSaveBtn?: boolean;
};

export const ModalWrapper: FC<ModalWrapperProps> = ({
  title,
  showSaveBtn = true,
  children,
  formId,
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
      {showSaveBtn && (
        <DialogFooter>
          <Button type="submit" form={formId ?? undefined}>
            {t("pages.explorer.endpoint.editor.form.plugins.saveBtn")}
          </Button>
        </DialogFooter>
      )}
    </DialogContent>
  );
};

type PluginSelectorProps = PropsWithChildren & {
  title: string;
};

export const PluginSelector: FC<PluginSelectorProps> = ({
  title,
  children,
}) => (
  <fieldset className="flex items-center gap-5">
    <label className="text-sm">{title}</label>
    {children}
  </fieldset>
);

export const PluginWrapper: FC<PropsWithChildren> = ({ children }) => (
  <Card className="flex flex-col gap-5 p-5" noShadow>
    {children}
  </Card>
);
