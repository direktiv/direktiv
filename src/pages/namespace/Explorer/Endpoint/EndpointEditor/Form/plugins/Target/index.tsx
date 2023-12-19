import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { ModalPluginSelector, ModalWrapper } from "../components/Modal";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { UseFormReturn, useWatch } from "react-hook-form";

import Button from "~/design/Button";
import { EndpointFormSchemaType } from "../../../schema";
import { InstantResponseForm } from "./InstantResponseForm";
import { Settings } from "lucide-react";
import { TargetFlowForm } from "./TargetFlowForm";
import { TargetFlowVarForm } from "./TargetFlowVarForm";
import { TargetNamespaceFileForm } from "./TargetNamespaceFileForm";
import { TargetNamespaceVarForm } from "./TargetNamespaceVarForm";
import { targetPluginTypes } from "../../../schema/plugins/target";

type TargetPluginFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
};

export const TargetPluginForm: FC<TargetPluginFormProps> = ({
  formControls,
}) => {
  const { control } = formControls;
  const values = useWatch({ control });
  const [dialogOpen, setDialogOpen] = useState(false);

  const currentType = values.plugins?.target?.type;
  const [selectedPlugin, setSelectedPlugin] = useState(currentType);

  const defaultInstantResponseConfig =
    currentType === targetPluginTypes.instantResponse
      ? values.plugins?.target?.configuration
      : undefined;

  const defaultTargetFlowConfig =
    currentType === targetPluginTypes.targetFlow
      ? values.plugins?.target?.configuration
      : undefined;

  const defaultTargetFlowVarConfig =
    currentType === targetPluginTypes.targetFlowVar
      ? values.plugins?.target?.configuration
      : undefined;

  const defaultTargetNamespaceFileConfig =
    currentType === targetPluginTypes.targetNamespaceFile
      ? values.plugins?.target?.configuration
      : undefined;

  const defaultTargetNamespaceVarConfig =
    currentType === targetPluginTypes.targetNamespaceVar
      ? values.plugins?.target?.configuration
      : undefined;

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <div className="flex items-center gap-3">
        Target plugin
        <DialogTrigger asChild>
          <Button icon variant="outline">
            <Settings /> {values.plugins?.target?.type ?? "no plugin set yet"}
          </Button>
        </DialogTrigger>
      </div>
      <ModalWrapper title="Configure Target plugin">
        <ModalPluginSelector title="Target plugin">
          <Select
            onValueChange={(e) => {
              setSelectedPlugin(e as typeof selectedPlugin);
            }}
            value={selectedPlugin}
          >
            <SelectTrigger variant="outline" className="grow">
              <SelectValue placeholder="please select a target plugin" />
            </SelectTrigger>
            <SelectContent>
              {Object.values(targetPluginTypes).map((pluginType) => (
                <SelectItem key={pluginType} value={pluginType}>
                  {pluginType}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </ModalPluginSelector>

        {selectedPlugin === targetPluginTypes.instantResponse && (
          <InstantResponseForm
            defaultConfig={defaultInstantResponseConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              formControls.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetPluginTypes.targetFlow && (
          <TargetFlowForm
            defaultConfig={defaultTargetFlowConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              formControls.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetPluginTypes.targetFlowVar && (
          <TargetFlowVarForm
            defaultConfig={defaultTargetFlowVarConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              formControls.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetPluginTypes.targetNamespaceFile && (
          <TargetNamespaceFileForm
            defaultConfig={defaultTargetNamespaceFileConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              formControls.setValue("plugins.target", configuration);
            }}
          />
        )}
        {selectedPlugin === targetPluginTypes.targetNamespaceVar && (
          <TargetNamespaceVarForm
            defaultConfig={defaultTargetNamespaceVarConfig}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              formControls.setValue("plugins.target", configuration);
            }}
          />
        )}
      </ModalWrapper>
    </Dialog>
  );
};
