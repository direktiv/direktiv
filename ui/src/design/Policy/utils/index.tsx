import { z } from "zod";

const _PathSchema = z.array(z.number()).min(1);
type Path = z.infer<typeof _PathSchema>;

export const getSum = (items: Path) =>
  items.reduce((sum, item) => sum + item, 0);
