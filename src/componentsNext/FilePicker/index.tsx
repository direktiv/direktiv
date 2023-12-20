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

import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { NodeSchemaType } from "~/api/tree/schema/node";
import { analyzePath } from "~/util/router/utils";
import { fileTypeToIcon } from "~/api/tree/utils";
import { twMergeClsx } from "~/util/helpers";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const convertFileToPath = (string?: string) =>
  analyzePath(string).parent?.absolute ?? "/";

const FilePicker = ({
  namespace,
  defaultPath,
  onChange,
  selectable,
}: {
  namespace?: string;
  defaultPath?: string;
  onChange?: (filePath: string) => void;
  selectable?: (node: NodeSchemaType) => boolean;
}) => {
  const [path, setPath] = useState(convertFileToPath(defaultPath));
  const [inputValue, setInputValue] = useState(defaultPath ? defaultPath : "");

  const { data, isError } = useNodeContent({
    path,
    namespace,
  });

  const { t } = useTranslation();

  const { parent, isRoot, segments } = analyzePath(path);

  const results = data?.children?.results ?? [];

  return (
    <ButtonBar>
      <Filepicker
        buttonText={t("components.filepicker.buttonText")}
        onClick={() => {
          setPath(convertFileToPath(inputValue));
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
        {isError && (
          <div>
            <FilepickerHeading>
              <div className="py-3">
                {t("components.filepicker.error.title", { path })}
              </div>
            </FilepickerHeading>
            <FilepickerList>
              <div
                onClick={() => {
                  setPath("/");
                }}
                className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
              >
                <FilepickerListItem icon={ArrowLeftToLineIcon}>
                  {t("components.filepicker.error.linkText")}
                </FilepickerListItem>
              </div>
            </FilepickerList>
          </div>
        )}
        {!isRoot && data && (
          <>
            <div
              onClick={() => {
                parent ? setPath(parent.absolute) : null;
              }}
              className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
            >
              <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
            </div>
            <FilepickerSeparator />
          </>
        )}
        <FilepickerList>
          {results.map((file) => {
            const isSelectable = selectable?.(file) ?? true;
            return (
              <Fragment key={file.name}>
                {file.type === "directory" ? (
                  <div
                    onClick={() => {
                      setPath(file.path);
                    }}
                    className="h-auto w-full cursor-pointer text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
                  >
                    <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                      {file.name}
                    </FilepickerListItem>
                  </div>
                ) : (
                  <FilepickerClose
                    className={twMergeClsx(
                      "h-auto w-full text-gray-11 hover:underline dark:text-gray-dark-11",
                      !isSelectable && "cursor-not-allowed opacity-70"
                    )}
                    disabled={!isSelectable}
                    onClick={() => {
                      setPath(file.parent);
                      setInputValue(file.path);
                      onChange?.(file.path);
                    }}
                  >
                    <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                      {file.name}
                    </FilepickerListItem>
                  </FilepickerClose>
                )}

                <FilepickerSeparator />
              </Fragment>
            );
          })}
        </FilepickerList>
      </Filepicker>
      <Input
        placeholder={t("components.filepicker.placeholder")}
        value={inputValue}
        className="w-80"
        onChange={(e) => {
          setInputValue(e.target.value);
          onChange?.(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default FilePicker;
