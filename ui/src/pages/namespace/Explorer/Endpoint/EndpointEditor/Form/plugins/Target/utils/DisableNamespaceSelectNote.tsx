import Alert from "~/design/Alert";
import { useTranslation } from "react-i18next";

export const DisableNamespaceSelectNote = () => {
  const { t } = useTranslation();
  return (
    <Alert variant="info">
      {t(
        "pages.explorer.endpoint.editor.form.plugins.target.disableNamespaceSelectNote"
      )}
    </Alert>
  );
};
