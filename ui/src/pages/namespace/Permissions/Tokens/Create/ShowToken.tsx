import { CheckCircle2, Eye, EyeOff, XCircle } from "lucide-react";
import { DialogFooter, DialogHeader, DialogTitle } from "~/design/Dialog";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import CopyButton from "~/design/CopyButton";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
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
  const [revealToken, setRevealToken] = useState(false);
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
          <InputWithButton>
            <Input
              value={token}
              readOnly
              type={revealToken ? "text" : "password"}
            />
            <Button
              variant="outline"
              onClick={() => setRevealToken(!revealToken)}
              icon
            >
              {revealToken ? <EyeOff /> : <Eye />}
            </Button>
          </InputWithButton>
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
