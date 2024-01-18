import { BadgeCheck, BadgeHelp } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const exampleService = `direktiv_api: "service/v1"
image: "redis"
scale: 1 # number of standby service replicas (optional)
size: "medium" # size of the image small, medium or large (optional)
cmd: "redis-server" # container's cmd string (optional)
envs: # list of environment variables (optional)
  - name: "MY_ENV_VAR"
    value: "env-var-value"
`;

const ServiceHelp = () => {
  const theme = useTheme();
  const { t } = useTranslation();
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" icon>
          <BadgeHelp />
        </Button>
      </PopoverTrigger>
      <PopoverContent asChild align="end" side="top">
        <div className="w-max p-5">
          <h3 className="mb-5 flex items-center gap-x-2 font-bold">
            <BadgeCheck className="h-5" />
            <div className="grow">
              {t("pages.explorer.service.editor.helpTitle")}
            </div>
            <CopyButton
              value={exampleService}
              buttonProps={{
                size: "sm",
              }}
            />
          </h3>
          <div className="flex h-[280px] w-[750px]">
            <Card className="flex grow p-4" noShadow>
              <Editor
                value={exampleService}
                theme={theme ?? undefined}
                options={{
                  readOnly: true,
                }}
              />
            </Card>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
};

export default ServiceHelp;
