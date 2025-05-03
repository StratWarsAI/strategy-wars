import { ScrollArea } from '@/components/ui/scroll-area';

export default function Container({
  children,
  scrollable = false
}: {
  children: React.ReactNode;
  scrollable?: boolean;
}) {
  return (
    <>
      {scrollable ? (
        <ScrollArea 
          className="h-[calc(100vh-52px)]"
          type="always"
        >
          <div className="p-4 md:px-8">{children}</div>
        </ScrollArea>
      ) : (
        <div className="h-full p-4 md:px-8">{children}</div>
      )}
    </>
  );
}