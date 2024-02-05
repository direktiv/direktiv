import { FC, PropsWithChildren } from "react";
import {
  VariablepickerError,
  VariablepickerMessage,
} from "~/design/VariablePicker";

import { useTranslation } from "react-i18next";

export const VariablePickerError: FC<PropsWithChildren> = ({ children }) => {
  const { t } = useTranslation();
  return (
    <VariablepickerError
      buttonText={t("components.workflowVariablepicker.buttonText")}
    >
      <VariablepickerMessage>{children}</VariablepickerMessage>
    </VariablepickerError>
  );
};
