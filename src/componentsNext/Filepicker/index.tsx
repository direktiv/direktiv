import { Breadcrumb, BreadcrumbRoot } from "~/design/Breadcrumbs";
import {
  Filepicker,
  FilepickerClose,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerSeparator,
} from "~/design/Filepicker";
import { FolderUp, Home } from "lucide-react";
import { Fragment, useState } from "react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { analyzePath } from "~/util/router/utils";
import { fileTypeToIcon } from "~/api/tree/utils";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const FilepickerMenu = ({
  namespace,
  defaultPath,
}: {
  namespace: string;
  defaultPath?: string;
}) => {
  const [path, setPath] = useState(defaultPath ? defaultPath : "/");

  const givenNamespace = namespace ? namespace : "";

  const { data, isFetched } = useNodeContent({
    path,
    givenNamespace,
  });
  const { parent, isRoot, segments } = analyzePath(path);

  const [file, setFile] = useState("");

  const { t } = useTranslation();

  if (!isFetched) return null;

  const results = data?.children?.results ?? [];

  return (
    <ButtonBar>
      <Filepicker
        buttonText={t("components.gatewayForms.Filepicker.buttonText")}
      >
        <FilepickerHeading>
          <BreadcrumbRoot className="py-3">
            <Breadcrumb
              noArrow
              onClick={() => {
                setPath("/");
              }}
              className="hover:underline"
            >
              <Home />
            </Breadcrumb>
            {segments.map((file) => {
              const isEmpty = file.absolute === "";

              if (isEmpty) return null;
              return (
                <Breadcrumb
                  key={file.relative}
                  onClick={() => {
                    setPath(file.absolute);
                  }}
                  className="hover:underline"
                >
                  {file.relative}
                </Breadcrumb>
              );
            })}
          </BreadcrumbRoot>
        </FilepickerHeading>
        <FilepickerSeparator />
        {!isRoot && (
          <Fragment>
            <FilepickerListItem icon={FolderUp}>
              <Button
                variant="link"
                onClick={() => {
                  parent ? setPath(parent.absolute) : null;
                }}
                className="h-auto p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent "
              >
                ..
              </Button>
            </FilepickerListItem>
            <FilepickerSeparator />
          </Fragment>
        )}
        <FilepickerList>
          {results.map((file) => (
            <Fragment key={file.name}>
              <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                {file.type === "directory" ? (
                  <Button
                    variant="link"
                    onClick={() => {
                      setPath(file.path);
                    }}
                    className="h-auto p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
                  >
                    {file.name}
                  </Button>
                ) : (
                  <FilepickerClose onClick={() => setFile(file.path)}>
                    {file.name}
                  </FilepickerClose>
                )}
              </FilepickerListItem>
              <FilepickerSeparator />
            </Fragment>
          ))}
        </FilepickerList>
      </Filepicker>
      <Input
        placeholder={t("components.gatewayForms.Filepicker.placeholder")}
        value={file}
        className="w-80"
        onChange={(e) => {
          setFile(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default FilepickerMenu;
