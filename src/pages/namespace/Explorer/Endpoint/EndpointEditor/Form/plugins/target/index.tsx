import { Controller, UseFormReturn, useWatch } from "react-hook-form";
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

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { EndpointFormSchemaType } from "../../../schema";
import { InstantResponseForm } from "./InstantResponseForm";
import { Settings } from "lucide-react";
import { targetPluginTypes } from "../../../schema/plugins/target";

type TargetPluginFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
};

export const TargetPluginForm: FC<TargetPluginFormProps> = ({
  formControls,
}) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const values = useWatch({
    control: formControls.control,
  });
  const { control } = formControls;

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <Card className="flex items-center gap-3 p-5">
        Target plugin
        <DialogTrigger asChild>
          <Button icon variant="outline">
            <Settings /> {values.plugins?.target?.type ?? "no plugin set yet"}
          </Button>
        </DialogTrigger>
      </Card>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure Target plugin</DialogTitle>
        </DialogHeader>
        <div className="my-3 flex flex-col gap-y-5">
          <Controller
            control={control}
            name="plugins.target.type"
            render={({ field }) => (
              <div className="flex flex-col gap-y-5">
                <fieldset className="flex items-center gap-5">
                  <label className="w-[150px] overflow-hidden text-right text-sm">
                    select a target plugin
                  </label>
                  <Select
                    /**
                     * TODO: this might not directly set the value, and more show which item is selected
                     * maybe we can use this and create a new component form all of this
                     */
                    onValueChange={field.onChange}
                    value={field.value}
                  >
                    <SelectTrigger variant="outline">
                      <SelectValue placeholder="please select a target plugin" />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.values(targetPluginTypes).map(
                        (targetPluginType) => (
                          <SelectItem
                            key={targetPluginType}
                            value={targetPluginType}
                          >
                            {targetPluginType}
                          </SelectItem>
                        )
                      )}
                    </SelectContent>
                  </Select>
                </fieldset>
              </div>
            )}
          />
          <Controller
            control={control}
            name="plugins.target"
            render={({ field: { value } }) => {
              if (value.type === targetPluginTypes.instantResponse) {
                return (
                  <InstantResponseForm
                    defaultConfig={value.configuration}
                    onSubmit={(configuration) => {
                      setDialogOpen(false);
                      formControls.setValue("plugins.target", configuration);
                    }}
                  />
                );
              }
              if (value.type === targetPluginTypes.targetFlow) {
                return <div>target flow</div>;
              }
              return <div>no plugin selected</div>;
            }}
          />
        </div>
      </DialogContent>
    </Dialog>
  );
};
