import { Fragment } from "react/jsx-runtime";

export const Preview = ({
  path,
  placeholders,
}: {
  path: string[];
  placeholders: string[];
}) => {
  const previewLength = Math.max(placeholders.length, path.length);

  return (
    <>
      {"{{"}
      {Array.from({ length: previewLength }, (_, index) => (
        <Fragment key={index}>
          {path[index] ? (
            <span className="text-gray-12 dark:text-gray-8">{path[index]}</span>
          ) : (
            <span className="italic text-gray-10">{placeholders[index]}</span>
          )}
          {index < previewLength - 1 && <span className="text-gray-10">.</span>}
        </Fragment>
      ))}
      {"}}"}
    </>
  );
};
