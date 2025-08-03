import { z } from "zod";

export const taskSchema = z.object({
  title: z.string()
    .min(1, "Task title is required")
    .min(3, "Task title must be at least 3 characters")
    .max(100, "Task title must not exceed 100 characters"),
  priority: z.enum(['none', 'low', 'medium', 'high']).default('none'),
  dueDate: z.string().optional().refine((date) => {
    if (!date) return true;
    const parsedDate = new Date(date);
    return !isNaN(parsedDate.getTime());
  }, "Invalid date format"),
  notes: z.string().max(500, "Notes must not exceed 500 characters").optional(),
});

export type TaskFormData = z.infer<typeof taskSchema>;