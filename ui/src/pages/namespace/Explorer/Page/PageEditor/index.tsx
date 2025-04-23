import {
  ArrowDownFromLine,
  ArrowDownToLine,
  Save,
  Table,
  Text,
} from "lucide-react";
import { Dialog, DialogContent } from "~/design/Dialog";
import { DragAndDropPreview, getElementComponent } from "./PreviewElements";
import { FC, Fragment, useState } from "react";
import {
  LayoutSchemaType,
  PageElementContentSchemaType,
  PageFormSchemaType,
} from "./schema";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";
import { decode, encode } from "js-base64";
import { defaultConfig, serializePageFile } from "./utils";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import DeleteModal from "./modals/Delete";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { DraggableElement } from "~/design/DragAndDropEditor/DraggableElement";
import { DroppableElement } from "~/design/DragAndDropEditor/DroppableElement";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
import EditModal from "./modals/Edit";
import { FileSchemaType } from "~/api/files/schema";
import { Form } from "./Form";
import FormErrors from "~/components/FormErrors";
import NavigationBlocker from "~/components/NavigationBlocker";
import { ScrollArea } from "~/design/ScrollArea";
import { jsonToYaml } from "../../utils";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type PageEditorProps = {
  data: FileSchemaType;
};

const PageEditor: FC<PageEditorProps> = ({ data }) => {
  const { t } = useTranslation();
  const fileContentFromServer = decode(data.data);
  const [pageConfig, pageConfigError] = serializePageFile(
    fileContentFromServer
  );
  const { mutate: updateRoute, isPending } = useUpdateFile();

  const [selectedDialog, setSelectedDialog] = useState<string>("edit");
  const [selectedElement, setSelectedElement] = useState<number>(0);

  const [dialogOpen, setDialogOpen] = useState<boolean>(false);

  const defaultLayout = defaultConfig.layout;

  const [layout, setLayout] = useState<LayoutSchemaType>(
    pageConfig?.layout ?? defaultLayout
  );

  const onMove = (
    name: string,
    target: string,
    position: "before" | "after" | undefined
  ) => {
    if (target) {
      const defaultTableData = [
        {
          header: "Table Header 1",
          cell: "- no data -",
        },
      ];

      const newElement = {
        name,
        hidden: false,
        content: {
          type: name,
          content: name === "Table" ? defaultTableData : `Placeholder ${name} `,
        } as unknown as PageElementContentSchemaType,
        preview: `Placeholder ${name} `,
      };
      const newLayout = [...layout];

      if (position === "before") {
        newLayout.splice(Number(target), 0, newElement);
      } else {
        if (position === "after") {
          newLayout.splice(Number(target + 1), 0, newElement);
        } else {
          newLayout[Number(target)] = newElement;
        }
      }

      setLayout(newLayout);
    }
  };

  const save = (value: PageFormSchemaType) => {
    const newValue = { ...value, layout };
    const toSave = jsonToYaml(newValue);

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
                              "pages.explorer.page.editor.form.serialisationError"
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
                          <FormErrors errors={errors} className="mb-5" />
                          {formMarkup}
                          <div className="pt-8 pb-4">
                            <h2 className="flex text-sm">
                              <ArrowDownFromLine size={20} className="mr-2" />
                              {t("pages.explorer.page.editor.form.sectionOne")}
                            </h2>
                          </div>
                          <Tabs defaultValue="data" className="w-full">
                            <TabsList variant="boxed">
                              <TabsTrigger variant="boxed" value="data">
                                {t(
                                  "pages.explorer.page.editor.form.tabsPageOne"
                                )}
                              </TabsTrigger>
                              <TabsTrigger variant="boxed" value="static">
                                {t(
                                  "pages.explorer.page.editor.form.tabsPageTwo"
                                )}
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
                                <DraggableElement icon={Text} name="Text" />
                              </Card>
                            </TabsContent>
                          </Tabs>

                          <div className="pt-8 pb-4">
                            <h2 className="flex text-sm">
                              <ArrowDownToLine size={20} className="mr-2" />
                              {t("pages.explorer.page.editor.form.sectionTwo")}
                            </h2>
                          </div>
                          <Card
                            noShadow
                            className="relative w-full bg-gray-2 dark:bg-gray-dark-2 rounded-md p-4"
                          >
                            {!layout.length && (
                              <DroppableSeparator
                                id={`${String(0)}-before`}
                                position="before"
                              />
                            )}
                            {layout.map((element, index) => {
                              const isLastListItem =
                                index === layout.length - 1;
                              return (
                                <Fragment key={index}>
                                  {isLastListItem ? (
                                    <>
                                      <DroppableSeparator
                                        id={`${index}-before`}
                                        position="before"
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
                                        id={`${index}-after`}
                                        position="after"
                                      />
                                    </>
                                  ) : (
                                    <>
                                      <DroppableSeparator
                                        id={`${index}-before`}
                                        position="before"
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
                          </Card>
                        </div>
                      )}
                    </Card>
                    <Card className="flex grow p-4 max-lg:h-[500px] bg-gray-2 dark:bg-gray-dark-2">
                      <Card
                        noShadow
                        className="ring-0 size-full bg-white dark:bg-black p-4"
                      >
                        {layout.map((element, index) => (
                          <DragAndDropPreview key={index}>
                            {getElementComponent(
                              element.name,
                              element.hidden,
                              element.content
                            )}
                          </DragAndDropPreview>
                        ))}
                      </Card>
                    </Card>
                  </div>
                  <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
                    {isDirty && (
                      <div className="text-sm text-gray-8 dark:text-gray-dark-8">
                        <span className="text-center">
                          {t("pages.explorer.page.editor.unsavedNote")}
                        </span>
                      </div>
                    )}
                    <Button
                      variant={isDirty ? "primary" : "outline"}
                      disabled={disableButton}
                      type="submit"
                    >
                      <Save />
                      {t("pages.explorer.page.editor.saveBtn")}
                    </Button>
                  </div>
                </div>
              </form>
            );
          }}
        </Form>
      </DndContext>

      <DialogContent className="overflow-auto min-w-[950px]">
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
