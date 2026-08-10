import { createBrowserClient } from "@supabase/ssr";

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || 'https://tmxvmtxddffemvtzwoyi.supabase.co';
const supabaseKey = process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY || 'sb_publishable_0mgYdtyTDDOsw4F1zHejcw_csWzII00';

export const createClient = () =>
  createBrowserClient(
    supabaseUrl!,
    supabaseKey!,
  );
