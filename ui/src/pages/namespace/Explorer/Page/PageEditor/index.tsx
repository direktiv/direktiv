import {
  DroppableElement,
  Placeholder,
} from "~/design/DragAndDropEditor/DroppableElement";
import { FC, ReactNode, useState } from "react";
import { Image, List, Save, TableIcon } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { decode, encode } from "js-base64";

import Alert from "~/design/Alert";
import Avatar from "~/design/Avatar";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { DragAndDropPreview } from "~/design/DragAndDropEditor";
import { Draggable } from "~/design/DragAndDropEditor/DraggableElement";
import { FileSchemaType } from "~/api/files/schema";
import { Form } from "./Form";
import FormErrors from "~/components/FormErrors";
import NavigationBlocker from "~/components/NavigationBlocker";
import { PageFormSchemaType } from "./schema";
import { ScrollArea } from "~/design/ScrollArea";
import { jsonToYaml } from "../../utils";
import { serializePageFile } from "./utils";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type PageEditorProps = {
  data: NonNullable<FileSchemaType>;
};

type person = {
  name?: string;
  email?: string;
  title?: string;
  role?: string;
};

type TableProps = {
  columns?: number;
  rows?: number;
  people?: person[];
};

const DefaultTable: FC<TableProps> = () => (
  <Table>
    <TableHead>
      <TableRow>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>
          <span className="sr-only">Edit</span>
        </TableHeaderCell>
      </TableRow>
    </TableHead>
    <TableBody>
      <TableRow>
        <TableCell>...</TableCell>
        <TableCell>...</TableCell>
        <TableCell>...</TableCell>
        <TableCell>...</TableCell>
      </TableRow>
    </TableBody>
  </Table>
);

const DefaultList: FC = () => (
  <ul className="list-disc">
    <li>...</li>
    <li>...</li>
    <li>...</li>
  </ul>
);

const getElementComponent = (element: string) => {
  switch (element) {
    case "Image":
      return <Avatar />;
    case "Table":
      return <DefaultTable />;
    case "List":
      return <DefaultList />;
    default:
      return <div>emptyyyy</div>;
  }
};

const PageEditor: FC<PageEditorProps> = ({ data }) => {
  const { t } = useTranslation();
  const fileContentFromServer = decode(data.data ?? "");
  const [pageConfig, pageConfigError] = serializePageFile(
    fileContentFromServer
  );
  const { mutate: updateRoute, isPending } = useUpdateFile();

  //  check for pageConfig === undefined
  //  pageConfig === undefined ? "" : pageConfig.layout[0].element

  const [elementName, setElementName] = useState<string>("Table");
  const [content, setContent] = useState<ReactNode>(<div></div>);

  //const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  const [header, setHeader] = useState<ReactNode>(<div></div>);
  const [footer, setFooter] = useState<ReactNode>(<div></div>);

  if (!header) {
    header !== "<div>" ? setHeader(<div></div>) : setHeader(<div></div>);
    footer !== "<div>" ? setFooter(<div></div>) : setFooter(<div></div>);
  }

  const onMove = (element: string, target: string) => {
    if (target) {
      const elementComponent = getElementComponent(element);
      setContent(elementComponent);
      setElementName(element);
    }
  };

  const save = (value: PageFormSchemaType) => {
    const toSave = jsonToYaml(value);
    updateRoute({
      path: data.path,
      payload: { data: encode(toSave) },
    });
  };

  const defaultConfig: PageFormSchemaType = {
    layout: [
      {
        element: "header",
      },
    ],
    direktiv_api: "page/v1",
    path: undefined,
  };

  return (
    <DndContext onMove={onMove}>
      <Form defaultConfig={pageConfig ?? defaultConfig} onSave={save}>
        {({
          formControls: {
            formState: { errors },
            handleSubmit,
          },
          formMarkup,
        }) => {
          // const preview = jsonToYaml(values);
          //const parsedOriginal = pageConfig && jsonToYaml(pageConfig);
          // const filehasChanged = preview !== parsedOriginal;
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
                  <Card className="p-5 lg:h-[calc(100vh-15.5rem)]">
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
                        <Card className="h-full bg-gray-1 p-4">
                          <Draggable name="Image">
                            <Button asChild variant="outline" size="lg">
                              <div className="w-28 bg-white">
                                <Image size={16} />
                                Image
                              </div>
                            </Button>
                          </Draggable>
                          <Draggable name="Table">
                            <Button asChild variant="outline" size="lg">
                              <div className="w-28 bg-white ">
                                <TableIcon size={16} />
                                Table
                              </div>
                            </Button>
                          </Draggable>
                          <Draggable name="List">
                            <Button asChild variant="outline" size="lg">
                              <div className="w-28 bg-white">
                                <List size={16} />
                                List
                              </div>
                            </Button>
                          </Draggable>
                        </Card>
                        <Placeholder name="Header" />
                        <DroppableElement
                          droppedElementName={elementName}
                        ></DroppableElement>
                        <FormErrors errors={errors} className="mb-5" />
                        <Placeholder name="Footer" />
                      </div>
                    )}
                  </Card>
                  <Card className="flex grow p-4 max-lg:h-[500px]">
                    <DragAndDropPreview>{content}</DragAndDropPreview>
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
  );
};

export default PageEditor;
