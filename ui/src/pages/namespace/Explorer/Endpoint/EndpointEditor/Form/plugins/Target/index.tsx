import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { UseFormReturn, useWatch } from "react-hook-form";
import {
  availablePlugins,
  targetPluginTypes,
} from "../../../schema/plugins/target";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { EndpointFormSchemaType } from "../../../schema";
import { InstantResponseForm } from "./InstantResponseForm";
import { ModalWrapper } from "~/components/ModalWrapper";
import { PluginSelector } from "../components/PluginSelector";
import { Settings } from "lucide-react";
import { TableHeader } from "../components/PluginsTable";
import { TargetEventForm } from "./TargetEventForm";
import { TargetFlowForm } from "./TargetFlowForm";
import { TargetFlowVarForm } from "./TargetFlowVarForm";
import { TargetNamespaceFileForm } from "./TargetNamespaceFileForm";
import { TargetNamespaceVarForm } from "./TargetNamespaceVarForm";
import { useTranslation } from "react-i18next";

type TargetPluginFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
};

export const TargetPluginForm: FC<TargetPluginFormProps> = ({
  form,
  onSave,
}) => {
  const { control, handleSubmit: parentSubmit } = form;
  const values = useWatch({ control });
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);

  const currentType = values.plugins?.target?.type;
  const [selectedPlugin, setSelectedPlugin] = useState(currentType);

  const {
    instantResponse,
    targetFlow,
    targetFlowVar,
    targetNamespaceFile,
    targetNamespaceVar,
    targetEvent,
  } = targetPluginTypes;

  const currentConfiguration = values.plugins?.target;

  const currentInstantResponseConfig =
    currentConfiguration?.type === instantResponse.name
      ? currentConfiguration.configuration
      : undefined;

  const currentTargetEventConfig =
    currentConfiguration?.type === targetEvent.name
      ? currentConfiguration.configuration
      : undefined;

  const currentTargetFlowConfig =
    currentConfiguration?.type === targetFlow.name
      ? currentConfiguration.configuration
      : undefined;

  const currentTargetFlowVarConfig =
    currentConfiguration?.type === targetFlowVar.name
      ? currentConfiguration.configuration
      : undefined;

  const currentTargetNamespaceFileConfig =
    currentConfiguration?.type === targetNamespaceFile.name
      ? currentConfiguration.configuration
      : undefined;

  const currentTargetNamespaceVarConfig =
    currentConfiguration?.type === targetNamespaceVar.name
      ? currentConfiguration.configuration
      : undefined;

  const formId = "targetPluginForm";

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <Card noShadow>
        <Table>
          <TableHeader
            title={t(
              "pages.explorer.endpoint.editor.form.plugins.target.table.headline"
            )}
          />
          <TableBody>
            <TableRow>
              <TableCell colSpan={2}>
                {values.plugins?.target?.type ? (
                  <DialogTrigger asChild>
                    <div className="cursor-pointer">
                      {t(
                        `pages.explorer.endpoint.editor.form.plugins.target.types.${values.plugins?.target?.type}`
                      )}
                    </div>
                  </DialogTrigger>
                ) : (
                  <div className="text-center">
                    <DialogTrigger asChild>
                      <Button icon variant="outline" size="sm">
                        <Settings />
                        {t(
                          "pages.explorer.endpoint.editor.form.plugins.target.table.addButton"
                        )}
                      </Button>
                    </DialogTrigger>
                  </div>
                )}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>

      <ModalWrapper
        formId={formId}
        showSaveBtn={!!selectedPlugin}
        title={t(
          "pages.explorer.endpoint.editor.form.plugins.target.modal.headline"
        )}
        onCancel={() => {
          setDialogOpen(false);
        }}
      >
        <PluginSelector
          title={t(
            "pages.explorer.endpoint.editor.form.plugins.target.modal.label"
          )}
        >
          <Select
            onValueChange={(e) => {
              setSelectedPlugin(e as typeof selectedPlugin);
            }}
            value={selectedPlugin}
          >
            <SelectTrigger variant="outline" className="grow">
              <SelectValue
                placeholder={t(
                  "pages.explorer.endpoint.editor.form.plugins.target.modal.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {availablePlugins.map(({ name: pluginName }) => (
                <SelectItem key={pluginName} value={pluginName}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.target.types.${pluginName}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>

        {selectedPlugin === instantResponse.name && (
          <InstantResponseForm
            formId={formId}
            defaultConfig={currentInstantResponseConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              form.setValue("plugins.target", configuration);
              parentSubmit(onSave)();
            }}
          />
        )}
        {selectedPlugin === targetFlow.name && (
          <TargetFlowForm
            formId={formId}
            defaultConfig={currentTargetFlowConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              form.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetFlowVar.name && (
          <TargetFlowVarForm
            formId={formId}
            defaultConfig={currentTargetFlowVarConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              form.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetNamespaceFile.name && (
          <TargetNamespaceFileForm
            formId={formId}
            defaultConfig={currentTargetNamespaceFileConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              form.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetNamespaceVar.name && (
          <TargetNamespaceVarForm
            formId={formId}
            defaultConfig={currentTargetNamespaceVarConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              form.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetEvent.name && (
          <TargetEventForm
            formId={formId}
            defaultConfig={currentTargetEventConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              form.setValue("plugins.target", configuration);
            }}
          />
        )}
      </ModalWrapper>
    </Dialog>
  );
};
