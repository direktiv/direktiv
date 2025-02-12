import {
  ArrowDownFromLine,
  ArrowDownToLine,
  Image,
  Save,
  Table,
  Text,
} from "lucide-react";
import { Dialog, DialogContent } from "~/design/Dialog";
import { DragAndDropPreview, getElementComponent } from "./PreviewElements";
import {
  DroppableElement,
  NonDroppableElement,
} from "~/design/DragAndDropEditor/DroppableElement";
import { FC, Fragment, useState } from "react";
import {
  LayoutSchemaType,
  PageElementSchemaType,
  PageFormSchemaType,
} from "./schema";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";
import { decode, encode } from "js-base64";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import DeleteModal from "./modals/Delete";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { DraggableElement } from "~/design/DragAndDropEditor/DraggableElement";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
import EditModal from "./modals/Edit";
import { FileSchemaType } from "~/api/files/schema";
import FooterForm from "./modals/forms/Footer";
import { Form } from "./Form";
import FormErrors from "~/components/FormErrors";
import HeaderForm from "./modals/forms/Header";
import NavigationBlocker from "~/components/NavigationBlocker";
import { ScrollArea } from "~/design/ScrollArea";
import { jsonToYaml } from "../../utils";
import { serializePageFile } from "./utils";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type PageEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const PageEditor: FC<PageEditorProps> = ({ data }) => {
  const { t } = useTranslation();
  const fileContentFromServer = decode(data.data ?? "");
  const [pageConfig, pageConfigError] = serializePageFile(
    fileContentFromServer
  );
  const { mutate: updateRoute, isPending } = useUpdateFile();

  const [selectedDialog, setSelectedDialog] = useState<string>("edit");
  const [selectedElement, setSelectedElement] = useState<number>(0);

  const [dialogOpen, setDialogOpen] = useState<boolean>(false);

  const headerDefault: PageElementSchemaType = {
    name: "Header",
    hidden: true,
    content: "This is the header",
    preview: "This is the header",
  };

  const footerDefault: PageElementSchemaType = {
    name: "Footer",
    hidden: false,
    content: "This is the footer",
    preview: "This is the footer",
  };

  const placeholder1: PageElementSchemaType = {
    name: "Text",
    hidden: false,
    content: "This is a Text...",
    preview: "This is a Text...",
  };

  const placeholder2: PageElementSchemaType = {
    name: "Table",
    hidden: false,
    content: [{ header: "Example Header", cell: "unset" }],
    preview: "Placeholder Table",
  };

  const placeholder3: PageElementSchemaType = {
    name: "Text",
    hidden: true,
    content: "some more info about...",
    preview: "some more info about...",
  };

  const defaultConfig: PageFormSchemaType = {
    header: headerDefault,
    footer: footerDefault,
    layout: [placeholder1, placeholder2, placeholder3],
    direktiv_api: "page/v1",
    path: undefined,
  };

  const defaultLayout = defaultConfig.layout;

  const [layout, setLayout] = useState<LayoutSchemaType>(
    pageConfig?.layout ?? defaultLayout
  );

  const [header, setHeader] = useState<PageElementSchemaType>(headerDefault);
  const [footer, setFooter] = useState<PageElementSchemaType>(footerDefault);

  const updateHeader = (header: PageElementSchemaType) => {
    const newHeader = header;
    newHeader.hidden = !header.hidden;
    setHeader(newHeader);
  };

  const updateFooter = (footer: PageElementSchemaType) => {
    const newFooter = footer;
    newFooter.hidden = !footer.hidden;
    setFooter(newFooter);
  };

  const onMove = (name: string, target: string) => {
    if (target) {
      const newElement = {
        name,
        hidden: false,
        content: `Placeholder ${name} `,
        preview: `Placeholder ${name} `,
      };
      const newLayout = [...layout];

      if (target.includes("A")) {
        target = target.slice(0, -1);
        newLayout.splice(Number(target), 0, newElement);
      } else {
        if (target.includes("B")) {
          target = target.slice(0, -1);
          newLayout.splice(Number(target + 1), 0, newElement);
        } else {
          newLayout[Number(target)] = newElement;
        }
      }

      setLayout(newLayout);
    }
  };

  const save = (value: PageFormSchemaType) => {
    const toSave = jsonToYaml(value);

    updateRoute({
      path: data.path,
      payload: { data: encode(toSave) },
    });
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DndContext onMove={onMove}>
        <Form defaultConfig={pageConfig ?? defaultConfig} onSave={save}>
          {({
            formControls: {
              formState: { errors },
              handleSubmit,
            },
            formMarkup,
          }) => {
            const filehasChanged = false;

            const isDirty = !pageConfigError && filehasChanged;
            const disableButton = isPending || !!pageConfigError;

            return (
              <form
                onSubmit={handleSubmit(save)}
                className="relative flex-col gap-4 p-5"
              >
                {isDirty && <NavigationBlocker />}
                <div className="flex flex-col gap-4">
                  <div className="grid grow grid-cols-1 gap-5 lg:grid-cols-2">
                    <Card className="p-5 lg:h-[calc(100vh-15.5rem)] overflow-x-clip  overflow-y-auto">
                      {pageConfigError ? (
                        <div className="flex flex-col gap-5 lg:overflow-y-scroll">
                          <Alert variant="error">
                            {t(
                              "pages.explorer.endpoint.editor.form.serialisationError"
                            )}
                          </Alert>
                          <ScrollArea className="size-full whitespace-nowrap">
                            <pre className="grow text-sm text-primary-500">
                              {JSON.stringify(pageConfigError, null, 2)}
                            </pre>
                          </ScrollArea>
                        </div>
                      ) : (
                        <div>
                          {formMarkup}
                          <div className="pt-8 pb-4">
                            <h2 className="flex text-sm">
                              <ArrowDownFromLine size={20} className="mr-2" />
                              Select Elements
                            </h2>
                          </div>
                          <Tabs defaultValue="data" className="w-full">
                            <TabsList variant="boxed">
                              <TabsTrigger variant="boxed" value="data">
                                Data Collection
                              </TabsTrigger>
                              <TabsTrigger variant="boxed" value="static">
                                Static Elements
                              </TabsTrigger>
                            </TabsList>
                            <TabsContent value="data" asChild>
                              <Card
                                className="p-4 text-sm bg-gray-2 dark:bg-gray-dark-2 flex row"
                                noShadow
                              >
                                <DraggableElement icon={Table} name="Table" />
                              </Card>
                            </TabsContent>
                            <TabsContent value="static" asChild>
                              <Card
                                className="p-4 text-sm bg-gray-2 dark:bg-gray-dark-2 flex row"
                                noShadow
                              >
                                <DraggableElement icon={Image} name="Image" />
                                <DraggableElement icon={Text} name="Text" />
                              </Card>
                            </TabsContent>
                          </Tabs>

                          <div className="pt-8 pb-4">
                            <h2 className="flex text-sm">
                              <ArrowDownToLine size={20} className="mr-2" />
                              Place Elements
                            </h2>
                          </div>
                          <Card
                            noShadow
                            className="relative w-full bg-gray-2 dark:bg-gray-dark-2 rounded-md p-4"
                          >
                            <NonDroppableElement
                              name="Header"
                              preview={header.preview}
                              hidden={header.hidden}
                              onHide={() => updateHeader(header)}
                              onEdit={() => {
                                setSelectedDialog("editHeader");
                                setDialogOpen(true);
                              }}
                            />
                            {!layout.length && (
                              <DroppableSeparator id={String(0) + "A"} />
                            )}
                            {layout.map((element, index) => {
                              const isLastListItem =
                                index === layout.length - 1;
                              return (
                                <Fragment key={index}>
                                  {isLastListItem ? (
                                    <>
                                      <DroppableSeparator
                                        id={String(index) + "A"}
                                      />
                                      <DroppableElement
                                        id={String(index)}
                                        name={element.name}
                                        preview={element.preview}
                                        hidden={element.hidden}
                                        onHide={() => {
                                          const newLayout = [...layout];
                                          const newElement = {
                                            ...element,
                                            hidden: !element.hidden,
                                          };
                                          newLayout.splice(
                                            index,
                                            1,
                                            newElement
                                          );
                                          setLayout(newLayout);
                                        }}
                                        setSelectedDialog={(dialogType) => {
                                          setSelectedDialog(dialogType);
                                          setSelectedElement(index);
                                          setDialogOpen(true);
                                        }}
                                      />
                                      <DroppableSeparator
                                        id={String(index) + "B"}
                                      />
                                    </>
                                  ) : (
                                    <>
                                      <DroppableSeparator
                                        id={String(index) + "A"}
                                      />
                                      <DroppableElement
                                        id={String(index)}
                                        name={element.name}
                                        preview={element.preview}
                                        hidden={element.hidden}
                                        onHide={() => {
                                          const newLayout = [...layout];
                                          const newElement = {
                                            ...element,
                                            hidden: !element.hidden,
                                          };
                                          newLayout.splice(
                                            index,
                                            1,
                                            newElement
                                          );
                                          setLayout(newLayout);
                                        }}
                                        setSelectedDialog={(dialogType) => {
                                          setSelectedDialog(dialogType);
                                          setSelectedElement(index);
                                          setDialogOpen(true);
                                        }}
                                      />
                                    </>
                                  )}
                                </Fragment>
                              );
                            })}
                            <FormErrors errors={errors} className="mb-5" />

                            <NonDroppableElement
                              name="Footer"
                              preview={footer.preview}
                              hidden={footer.hidden}
                              onHide={() => updateFooter(footer)}
                              onEdit={() => {
                                setSelectedDialog("editFooter");
                                setDialogOpen(true);
                              }}
                            />
                          </Card>
                        </div>
                      )}
                    </Card>
                    <Card className="flex grow p-4 max-lg:h-[500px] bg-gray-2 dark:bg-gray-dark-2">
                      <Card
                        noShadow
                        className="ring-0 size-full bg-white dark:bg-black p-4"
                      >
                        <DragAndDropPreview>
                          {getElementComponent(
                            header.name,
                            header.hidden,
                            header.content
                          )}
                        </DragAndDropPreview>
                        {layout.map((element, index) => (
                          <DragAndDropPreview key={index}>
                            {getElementComponent(
                              element.name,
                              element.hidden,
                              element.content
                            )}
                          </DragAndDropPreview>
                        ))}
                        <DragAndDropPreview>
                          {getElementComponent(
                            footer.name,
                            footer.hidden,
                            footer.content
                          )}
                        </DragAndDropPreview>
                      </Card>
                    </Card>
                  </div>
                  <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
                    {isDirty && (
                      <div className="text-sm text-gray-8 dark:text-gray-dark-8">
                        <span className="text-center">
                          {t("pages.explorer.endpoint.editor.unsavedNote")}
                        </span>
                      </div>
                    )}
                    <Button
                      variant={isDirty ? "primary" : "outline"}
                      disabled={disableButton}
                      type="submit"
                    >
                      <Save />
                      {t("pages.explorer.endpoint.editor.saveBtn")}
                    </Button>
                  </div>
                </div>
              </form>
            );
          }}
        </Form>
      </DndContext>

      <DialogContent className="overflow-auto min-w-[950px]">
        {selectedDialog === "editHeader" && (
          <HeaderForm
            header={header}
            onEdit={(newHeader) => setHeader(newHeader)}
            close={() => {
              setDialogOpen(false);
            }}
          />
        )}
        {selectedDialog === "editFooter" && (
          <FooterForm
            footer={footer}
            onEdit={(newFooter) => setFooter(newFooter)}
            close={() => {
              setDialogOpen(false);
            }}
          />
        )}
        {selectedDialog === "edit" && (
          <EditModal
            onChange={() => setSelectedElement(0)}
            layout={layout}
            success={(newLayout) => {
              setLayout(newLayout);
            }}
            pageElementID={selectedElement}
            close={() => {
              setDialogOpen(false);
            }}
          />
        )}
        {selectedDialog === "delete" && (
          <DeleteModal
            layout={layout}
            success={(newLayout) => {
              setLayout(newLayout);
            }}
            pageElementID={selectedElement}
            close={() => {
              setDialogOpen(false);
            }}
          />
        )}
      </DialogContent>
    </Dialog>
  );
};

export default PageEditor;
