import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Code } from "lucide-react";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const RoutePreview = ({ fileContent }: { fileContent: string }) => {
  const theme = useTheme();
  const { t } = useTranslation();
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" icon>
          <Code />
        </Button>
      </PopoverTrigger>
      <PopoverContent asChild align="end" side="top">
        <div className="w-max p-5">
          <h3 className="mb-5 flex items-center gap-x-2 font-bold">
            <Code className="h-5" />
            <div className="grow">
              {t("pages.explorer.endpoint.editor.previewTitle")}
            </div>
            <CopyButton
              value={fileContent}
              buttonProps={{
                size: "sm",
              }}
            />
          </h3>
          <div className="flex h-[500px] w-[650px]">
            <Card className="flex  grow p-4" noShadow>
              <Editor
                value={fileContent}
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

export default RoutePreview;
