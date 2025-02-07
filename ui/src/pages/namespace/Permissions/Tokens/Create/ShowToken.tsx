import { CheckCircle2, XCircle } from "lucide-react";
import { DialogFooter, DialogHeader, DialogTitle } from "~/design/Dialog";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import CopyButton from "~/design/CopyButton";
import PasswordInput from "~/pages/namespace/Gateway/Consumers/Table/Row/PasswordInput";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const ShowToken = ({
  token,
  onCloseClicked,
}: {
  token: string;
  onCloseClicked: () => void;
}) => {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <CheckCircle2 /> {t("pages.permissions.tokens.create.success.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Alert className="mb-5" variant="warning">
          {t("pages.permissions.tokens.create.success.description")}
        </Alert>
        <div className="flex gap-3">
          <PasswordInput password={token} />
          <CopyButton
            value={token}
            buttonProps={{
              variant: "outline",
              className: "w-60",
              onClick: () => setCopied(true),
            }}
          >
            {(copied) =>
              copied
                ? t("pages.permissions.tokens.create.success.copied")
                : t("pages.permissions.tokens.create.success.copy")
            }
          </CopyButton>
        </div>
      </div>
      <DialogFooter>
        <Button type="button" disabled={!copied} onClick={onCloseClicked}>
          <XCircle />
          {t("pages.permissions.tokens.create.success.close")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default ShowToken;
