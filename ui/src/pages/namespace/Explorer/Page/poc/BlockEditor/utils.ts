import { DirektivPagesType } from "../schema";

export const clonePage = (page: DirektivPagesType): DirektivPagesType =>
  structuredClone(page);
