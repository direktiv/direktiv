import { BookOpen, TerminalSquare } from "lucide-react";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import Input from "~/design/Input";
import { generateApiCommandTemplate } from "./utils";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const ApiCommands = () => {
  const namespaceFromUrl = useNamespace();
  const { path: pathFromUrl } = pages.explorer.useParams();
  const theme = useTheme();
  const [path, setPath] = useState(pathFromUrl ?? "");
  const [namespace, setNamespace] = useState(namespaceFromUrl ?? "");

  const apiCommandTemplates = generateApiCommandTemplate(namespace, path);
  const apiCommands = apiCommandTemplates.map((t) => t.key); // TODO: memoize

  const [selectedCommand, setSelectedCommand] = useState(apiCommands[0]);

  console.log("🚀", selectedCommand);

  const { t } = useTranslation();

  if (!namespaceFromUrl) return null;
  if (!pathFromUrl) return null;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <TerminalSquare />
          {t("pages.explorer.workflow.apiCommands.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <div className="flex flex-col gap-y-5">
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[100px] text-right text-[14px]"
              htmlFor="namespace-name"
            >
              {t("pages.explorer.workflow.apiCommands.namespace")}
            </label>
            <Input
              id="namespace-name"
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
              placeholder={t(
                "pages.explorer.workflow.apiCommands.namespacePlaceholder"
              )}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[100px] text-right text-[14px]"
              htmlFor="workflow-name"
            >
              {t("pages.explorer.workflow.apiCommands.workflow")}
            </label>
            <Input
              id="workflow-name"
              value={path}
              onChange={(e) => setPath(e.target.value)}
              placeholder={t(
                "pages.explorer.workflow.apiCommands.workflowPlaceholder"
              )}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[100px] text-right text-[14px]"
              htmlFor="template"
            >
              {t("pages.explorer.tree.newWorkflow.templateLabel")}
            </label>
            <Select
              onValueChange={(value) => {
                const matchingWf = apiCommandTemplates.find(
                  (template) => template.key === value
                );
                if (matchingWf) {
                  setSelectedCommand(matchingWf.key);
                }
              }}
            >
              <SelectTrigger id="template" variant="outline" block>
                <SelectValue
                  defaultValue={selectedCommand}
                  placeholder={selectedCommand}
                />
              </SelectTrigger>
              <SelectContent>
                {apiCommands.map((command) => (
                  <SelectItem value={command} key={command}>
                    {t(`pages.explorer.workflow.apiCommands.labels.${command}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          {/* <div className="flex flex-col text-sm"> */}
          <Card
            className="grid grid-cols-2 p-4 text-sm"
            noShadow
            background="weight-1"
          >
            <Badge variant="success">POST</Badge>
            <code className="text-primary-500">
              http://localhost:3000/api/namespaces/stefan/tree/dir/test.yaml?op=router
            </code>
          </Card>
          <Card className="h-44 p-4" noShadow background="weight-1">
            <Editor
              value="workflowData"
              language="yaml"
              options={{}}
              onChange={(newData) => {
                // if (newData) {
                //   setWorkflowData(newData);
                //   setValue("fileContent", newData);
                // }
              }}
              theme={theme ?? undefined}
            />
          </Card>
          <Card className="h-44 p-4" noShadow background="weight-1">
            <Editor
              value="curl 'http://localhost:3000/api/namespaces/stefan/tree/dir/test.yaml?op=router' \
              -H 'direktiv-token: Qhxw6U3#6&hu^j' \
              -H 'sec-ch-ua-mobile: ?0' \"
              language="shell"
              options={{ readOnly: true }}
              theme={theme ?? undefined}
            />
          </Card>
        </div>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Close</Button>
        </DialogClose>
        <Button variant="outline" asChild>
          <a
            href="https://docs.direktiv.io/api/#all-endpoints"
            target="_blank"
            rel="noopener noreferrer"
          >
            <BookOpen />
            {t("pages.explorer.workflow.apiCommands.openDocs")}
          </a>
        </Button>
        <CopyButton
          value="instance"
          buttonProps={{
            variant: "outline",
            className: "w-40",
          }}
        >
          {(copied) =>
            copied
              ? t("pages.explorer.workflow.apiCommands.copyBtnCopied")
              : t("pages.explorer.workflow.apiCommands.copyBtn")
          }
        </CopyButton>
      </DialogFooter>
    </>
  );
};

export default ApiCommands;
