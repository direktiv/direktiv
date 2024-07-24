const yamlExtensions = [
  ".yaml", // this first one is the default
  ".yml",
] as const;

export const forceYamlFileExtension = (name: string) => {
  const newName = name.trim();
  if (yamlExtensions.some((extension) => newName.endsWith(extension))) {
    return newName;
  }
  return `${newName}${yamlExtensions[0]}`;
};

/**
 * Intended use: remove partial or full existing extension from a file name
 * before adding automatic extension to avoid duplicating it.
 * @param name
 * @param extension
 * @returns name without extension (or partial matches)
 */
export const stripFileExtension = (name: string, extension: string) => {
  const nameSegments = name.split(".");
  const partials = extension?.split(".");

  partials?.reverse().forEach((partial) => {
    if (nameSegments[nameSegments.length - 1] === partial) {
      nameSegments.pop();
    }
  });

  return nameSegments.join(".");
};

/**
 * Will add the provided extension to the name. Will strip (partial) matches
 * of the extension first to avoid duplicating them.
 * @param name
 * @param extension
 * @returns name with extension
 */
export const forceFileExtension = (name: string, extension: string) => {
  const baseName = stripFileExtension(name.trim(), extension);
  return `${baseName}${extension}`;
};
