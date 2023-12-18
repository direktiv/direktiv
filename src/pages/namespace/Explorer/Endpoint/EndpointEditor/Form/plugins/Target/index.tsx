import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { FC, useState } from "react";
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
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure Target plugin</DialogTitle>
        </DialogHeader>
        <div className="my-3 flex flex-col gap-y-5">
          <div className="flex flex-col gap-y-5">
            <fieldset className="flex items-center gap-5">
              <label className="w-[150px] overflow-hidden text-right text-sm">
                select a target plugin
              </label>
              <Select
                onValueChange={(e) => {
                  setSelectedPlugin(e as typeof selectedPlugin);
                }}
                value={selectedPlugin}
              >
                <SelectTrigger variant="outline">
                  <SelectValue placeholder="please select a target plugin" />
                </SelectTrigger>
                <SelectContent>
                  {Object.values(targetPluginTypes).map((targetPluginType) => (
                    <SelectItem key={targetPluginType} value={targetPluginType}>
                      {targetPluginType}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </fieldset>
          </div>

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
          {selectedPlugin === targetPluginTypes.targetNamespaceFile && null}
          {selectedPlugin === targetPluginTypes.targetNamespaceVar && null}
        </div>
      </DialogContent>
    </Dialog>
  );
};
