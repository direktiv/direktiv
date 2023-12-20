import { ArrowDown, ArrowUp, MoreVertical, Plus, Trash } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, useState } from "react";
import { ModalWrapper, PluginSelector } from "../components/Modal";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { UseFormReturn, useFieldArray } from "react-hook-form";
import {
  getJsInboundConfigAtIndex,
  getRequestConvertConfigAtIndex,
} from "../utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { EndpointFormSchemaType } from "../../../schema";
import { InboundPluginFormSchemaType } from "../../../schema/plugins/inbound/schema";
import { JsInboundForm } from "./JsInboundForm";
import { RequestConvertForm } from "./RequestConvertForm";
import { inboundPluginTypes } from "../../../schema/plugins/inbound";
import { useTranslation } from "react-i18next";

type InboundPluginFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
};

export const InboundPluginForm: FC<InboundPluginFormProps> = ({ form }) => {
  const { t } = useTranslation();
  const { control } = form;
  const {
    append: addPlugin,
    remove: deletePlugin,
    move: movePlugin,
    update: editPlugin,
    fields,
  } = useFieldArray({
    control,
    name: "plugins.inbound",
  });
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const [selectedPlugin, setSelectedPlugin] =
    useState<InboundPluginFormSchemaType["type"]>();

  const { jsInbound, requestConvert } = inboundPluginTypes;

  const pluginsCount = fields.length;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <div className="flex items-center gap-3"></div>

      <Card noShadow>
        <Table>
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>{pluginsCount} inbound plugins</TableHeaderCell>
              <TableHeaderCell className="w-60 text-right">
                <DialogTrigger asChild>
                  <Button icon variant="outline" size="sm">
                    <Plus /> add inbound plugin
                  </Button>
                </DialogTrigger>
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {fields.map(({ id, type }, index, srcArray) => {
              const canMoveDown = index < srcArray.length - 1;
              const canMoveUp = index > 0;
              return (
                <TableRow
                  key={id}
                  className="cursor-pointer"
                  onClick={() => {
                    setSelectedPlugin(type);
                    setDialogOpen(true);
                    setEditIndex(index);
                  }}
                >
                  <TableCell>
                    {t(
                      `pages.explorer.endpoint.editor.form.plugins.inbound.types.${type}`
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => e.preventDefault()}
                          icon
                        >
                          <MoreVertical />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-40">
                        {canMoveDown && (
                          <DialogTrigger
                            className="w-full"
                            onClick={(e) => {
                              e.stopPropagation();
                              e.preventDefault();
                              movePlugin(index, index + 1);
                            }}
                          >
                            <DropdownMenuItem>
                              <ArrowDown className="mr-2 h-4 w-4" />
                              mode down
                            </DropdownMenuItem>
                          </DialogTrigger>
                        )}
                        {canMoveUp && (
                          <DialogTrigger
                            className="w-full"
                            onClick={(e) => {
                              e.stopPropagation();
                              e.preventDefault();
                              movePlugin(index, index - 1);
                            }}
                          >
                            <DropdownMenuItem>
                              <ArrowUp className="mr-2 h-4 w-4" />
                              mode up
                            </DropdownMenuItem>
                          </DialogTrigger>
                        )}

                        <DialogTrigger
                          className="w-full"
                          onClick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            deletePlugin(index);
                          }}
                        >
                          <DropdownMenuItem>
                            <Trash className="mr-2 h-4 w-4" />
                            delete
                          </DropdownMenuItem>
                        </DialogTrigger>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Card>

      <ModalWrapper
        title={
          editIndex === undefined
            ? t(
                "pages.explorer.endpoint.editor.form.plugins.inbound.headlineAdd"
              )
            : t(
                "pages.explorer.endpoint.editor.form.plugins.inbound.headlineEdit"
              )
        }
      >
        <PluginSelector
          title={t("pages.explorer.endpoint.editor.form.plugins.inbound.label")}
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
                  "pages.explorer.endpoint.editor.form.plugins.inbound.placeholder"
                )}
              />
            </SelectTrigger>
            <SelectContent>
              {Object.values(inboundPluginTypes).map((pluginType) => (
                <SelectItem key={pluginType} value={pluginType}>
                  {t(
                    `pages.explorer.endpoint.editor.form.plugins.inbound.types.${pluginType}`
                  )}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </PluginSelector>
        {selectedPlugin === requestConvert && (
          <RequestConvertForm
            defaultConfig={getRequestConvertConfigAtIndex(fields, editIndex)}
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
        {selectedPlugin === jsInbound && (
          <JsInboundForm
            defaultConfig={getJsInboundConfigAtIndex(fields, editIndex)}
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
