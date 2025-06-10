import { defineConfig } from "drizzle-kit";
 
export default defineConfig({
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DB_URL!,
  }
});
