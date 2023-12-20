import { ChevronDown, ChevronUp, Edit, Plus, Trash } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { ModalWrapper, PluginSelector } from "../components/Modal";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { UseFormReturn, useFieldArray } from "react-hook-form";
import {
  getBasicAuthConfigAtIndex,
  getGithubWebhookAuthConfigAtIndex,
  getKeyAuthConfigAtIndex,
} from "../utils";

import { AuthPluginFormSchemaType } from "../../../schema/plugins/auth/schema";
import { BasicAuthForm } from "./BasicAuthForm";
import Button from "~/design/Button";
import { EndpointFormSchemaType } from "../../../schema";
import { GithubWebhookAuthForm } from "./GithubWebhookAuthForm";
import { KeyAuthForm } from "./KeyAuthForm";
import { authPluginTypes } from "../../../schema/plugins/auth";
import { useTranslation } from "react-i18next";

type AuthPluginFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
};

export const AuthPluginForm: FC<AuthPluginFormProps> = ({ formControls }) => {
  const { t } = useTranslation();
  const { control } = formControls;
  const {
    append: addPlugin,
    remove: deletePlugin,
    move: movePlugin,
    update: editPlugin,
    fields,
  } = useFieldArray({
    control,
    name: "plugins.auth",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<AuthPluginFormSchemaType["type"]>();

  const pluginsCount = fields.length;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <div className="flex items-center gap-3">
        {pluginsCount} Auth plugins
        <DialogTrigger asChild>
          <Button icon variant="outline">
            <Plus /> add auth plugin
          </Button>
        </DialogTrigger>
      </div>
      {fields.map(({ id, type }, index, srcArray) => {
        const canMoveDown = index < srcArray.length - 1;
        const canMoveUp = index > 0;

        return (
          <div key={id} className="flex gap-2">
            {type}
            <Button
              variant="destructive"
              icon
              size="sm"
              onClick={() => {
                deletePlugin(index);
              }}
            >
              <Trash />
            </Button>
            <Button
              variant="outline"
              icon
              size="sm"
              disabled={!canMoveDown}
              onClick={() => {
                movePlugin(index, index + 1);
              }}
            >
              <ChevronDown />
            </Button>
            <Button
              variant="outline"
              icon
              size="sm"
              disabled={!canMoveUp}
              onClick={() => {
                movePlugin(index, index - 1);
              }}
            >
              <ChevronUp />
            </Button>
            <Button
              variant="outline"
              icon
              size="sm"
              onClick={() => {
                setSelectedPlugin(type);
                setDialogOpen(true);
                setEditIndex(index);
              }}
            >
              <Edit />
            </Button>
          </div>
        );
      })}
      <ModalWrapper
        title={
          editIndex === undefined
            ? t("pages.explorer.endpoint.editor.form.plugins.auth.headlineAdd")
            : t("pages.explorer.endpoint.editor.form.plugins.auth.headlineEdit")
        }
      >
        <PluginSelector
          title={t("pages.explorer.endpoint.editor.form.plugins.auth.label")}
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
                  "pages.explorer.endpoint.editor.form.plugins.auth.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {Object.values(authPluginTypes).map((pluginType) => (
                <SelectItem key={pluginType} value={pluginType}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.auth.types.${pluginType}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === authPluginTypes.basicAuth && (
          <BasicAuthForm
            defaultConfig={getBasicAuthConfigAtIndex(fields, editIndex)}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              if (editIndex === undefined) {
                addPlugin(configuration);
              } else {
                editPlugin(editIndex, configuration);
              }
              setEditIndex(undefined);
            }}
          />
        )}
        {selectedPlugin === authPluginTypes.keyAuth && (
          <KeyAuthForm
            defaultConfig={getKeyAuthConfigAtIndex(fields, editIndex)}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              if (editIndex === undefined) {
                addPlugin(configuration);
              } else {
                editPlugin(editIndex, configuration);
              }
              setEditIndex(undefined);
            }}
          />
        )}
        {selectedPlugin === authPluginTypes.githubWebhookAuth && (
          <GithubWebhookAuthForm
            defaultConfig={getGithubWebhookAuthConfigAtIndex(fields, editIndex)}
            onSubmit={(configuration) => {
              setDialogOpen(false);
              if (editIndex === undefined) {
                addPlugin(configuration);
              } else {
                editPlugin(editIndex, configuration);
              }
              setEditIndex(undefined);
            }}
          />
        )}
      </ModalWrapper>
    </Dialog>
  );
};
