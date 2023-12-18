import { ArrowLeftToLineIcon, FolderUp, Home } from "lucide-react";
import { Breadcrumb, BreadcrumbRoot } from "~/design/Breadcrumbs";
import {
  Filepicker,
  FilepickerClose,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerSeparator,
} from "~/design/Filepicker";
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
  onChange,
}: {
  namespace?: string;
  defaultPath?: string;
  onChange?: (filePath: string) => void;
}) => {
  const [path, setPath] = useState(defaultPath ? defaultPath : "/");
  const [inputValue, setInputValue] = useState(defaultPath ? defaultPath : "");

  const { data } = useNodeContent({
    path,
    namespace,
  });

  const { t } = useTranslation();

  const { parent, isRoot, segments } = analyzePath(path);

  const results = data?.children?.results ?? [];

  return (
    <div>
      <ButtonBar>
        <Filepicker
          buttonText={t("components.filepicker.buttonText")}
          onClick={() => {
            setPath(inputValue);
          }}
        >
          <FilepickerHeading>
            <BreadcrumbRoot className="py-3">
              <Breadcrumb
                noArrow
                onClick={() => {
                  setPath("/");
                }}
                className="h-5 hover:underline"
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
                    className="h-5 hover:underline"
                  >
                    {file.relative}
                  </Breadcrumb>
                );
              })}
            </BreadcrumbRoot>
          </FilepickerHeading>
          <FilepickerSeparator />
          {!data && (
            <div>
              <FilepickerHeading>
                <div className="py-3">The provided Path was not found.</div>
              </FilepickerHeading>
              <FilepickerList>
                <FilepickerListItem icon={ArrowLeftToLineIcon}>
                  <Button
                    onClick={() => {
                      setPath("/");
                    }}
                    variant="link"
                    className="h-auto p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
                  >
                    Go back to root directory
                  </Button>
                </FilepickerListItem>
              </FilepickerList>
            </div>
          )}
          {!isRoot && data && (
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
                    <FilepickerClose
                      onClick={() => {
                        setPath(file.parent);
                        setInputValue(file.path);
                        onChange?.(file.path);
                      }}
                    >
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
          placeholder={t("components.filepicker.placeholder")}
          value={inputValue}
          className="w-80"
          onChange={(e) => {
            setInputValue(e.target.value);
          }}
        />
      </ButtonBar>
    </div>
  );
};

export default FilepickerMenu;
