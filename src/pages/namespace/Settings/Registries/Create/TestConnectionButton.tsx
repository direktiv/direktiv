import { CheckCircle2, CircleDashed, LucideIcon, XCircle } from "lucide-react";
import { ComponentProps, useEffect, useState } from "react";

import Button from "~/design/Button";
import { useTestConnection } from "~/api/registries/mutate/testConnection";
import { useTranslation } from "react-i18next";

export const TestConnectionButton = ({
  isValid,
  getValues,
}: {
  isValid: boolean;
  getValues: (name: "url" | "user" | "password") => string;
}) => {
  const { t } = useTranslation();
  const [testSuccessful, setTestSuccessful] = useState<boolean | null>(null); // null = not tested yet

  useEffect(() => {
    // reset test status after 3 seconds
    if (testSuccessful !== null) {
      const timeout = setTimeout(() => {
        setTestSuccessful(null);
      }, 3000);
      return () => clearTimeout(timeout);
    }
  }, [testSuccessful]);

  const { mutate: testConnection, isLoading } = useTestConnection({
    onSuccess: () => {
      setTestSuccessful(true);
    },
    onError: () => {
      setTestSuccessful(false);
    },
  });

  const onTestConnectionClick = () => {
    testConnection({
      url: getValues("url"),
      username: getValues("user"),
      password: getValues("password"),
    });
  };

  let variant: ComponentProps<typeof Button>["variant"] = "outline";
  let Icon: LucideIcon = CircleDashed;
  let label = t("pages.settings.registries.create.testConnectionBtn.label");

  if (testSuccessful === true) {
    variant = "primary";
    Icon = CheckCircle2;
    label = t("pages.settings.registries.create.testConnectionBtn.success");
  }
  if (testSuccessful === false) {
    variant = "destructive";
    Icon = XCircle;
    label = t("pages.settings.registries.create.testConnectionBtn.error");
  }

  return (
    <Button
      data-testid="btn-test-connection"
      onClick={onTestConnectionClick}
      loading={isLoading}
      disabled={!isValid || isLoading}
      type="button"
      variant={variant}
    >
      {!isLoading && <Icon />}
      {label}
    </Button>
  );
};
