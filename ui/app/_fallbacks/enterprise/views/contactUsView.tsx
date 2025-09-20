import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface Props {
	className?: string;
	icon: React.ReactNode;
	title: string;
	description: string;
}

export default function ContactUsView({ icon, title, description, className }: Props) {
	return (
		<div className={cn("flex flex-col items-center justify-center gap-4 min-h-[80vh] text-center", className)}>
			<div className="text-muted-foreground">{icon}</div>
			<div className="flex flex-col gap-1">
				<div className="text-muted-foreground text-xl font-medium">{title}</div>
				<div className="text-muted-foreground text-sm font-normal max-w-[600px] mt-2">{description}</div>
                <Button className="w-[200px] mx-auto mt-6">Book a demo</Button>
			</div>
		</div>
	);
}
