'use client';

import {
  Home,
  LayoutDashboard,
  MapPin,
  ClipboardList,
  Settings,
  Users,
  Building2,
  Calendar,
  FileText,
  History,
  Activity,
  Sliders,
  ChevronLeft,
  ChevronRight,
  Bell,
  Search,
  Menu,
  X,
  Plus,
  Minus,
  Edit,
  Trash2,
  Save,
  Check,
  AlertCircle,
  AlertTriangle,
  Info,
  HelpCircle,
  Circle,
  Star,
  Heart,
  Eye,
  EyeOff,
  Lock,
  Unlock,
  Key,
  User,
  UserPlus,
  UserMinus,
  UserCog,
  Shield,
  ShieldCheck,
  ShieldAlert,
  Skull,
  Clock,
  Timer,
  RefreshCw,
  Download,
  Upload,
  ExternalLink,
  Link,
  Unlink,
  Mail,
  Phone,
  MessageSquare,
  Send,
  Inbox,
  Archive,
  Folder,
  FolderOpen,
  File,
  FilePlus,
  FileCheck,
  FileX,
  Copy,
  Clipboard,
  ClipboardCheck,
  Database,
  Server,
  Cloud,
  Wifi,
  WifiOff,
  Globe,
  Map,
  Navigation,
  Compass,
  Target,
  Crosshair,
  Zap,
  Power,
  Battery,
  BatteryCharging,
  Thermometer,
  Droplet,
  Sun,
  Moon,
  Sunrise,
  Sunset,
  Wind,
  Umbrella,
  Filter,
  SortAsc,
  SortDesc,
  List,
  Grid,
  Layers,
  Layout,
  Maximize,
  Minimize,
  Move,
  MoreHorizontal,
  MoreVertical,
  ChevronDown,
  ChevronUp,
  ChevronsLeft,
  ChevronsRight,
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ArrowDown,
  RotateCw,
  RotateCcw,
  Play,
  Pause,
  Square,
  SkipBack,
  SkipForward,
  Volume,
  Volume1,
  Volume2,
  VolumeX,
  Mic,
  MicOff,
  Camera,
  Image,
  Video,
  Film,
  Music,
  Headphones,
  Radio,
  Tv,
  Monitor,
  Smartphone,
  Tablet,
  Laptop,
  Printer,
  HardDrive,
  Cpu,
  Usb,
  Bluetooth,
  Cast,
  Share,
  Share2,
  Bookmark,
  Tag,
  Tags,
  Hash,
  AtSign,
  DollarSign,
  Percent,
  BarChart,
  BarChart2,
  PieChart,
  LineChart,
  TrendingUp,
  TrendingDown,
  Award,
  Gift,
  ShoppingCart,
  ShoppingBag,
  Package,
  Box,
  Truck,
  Plane,
  Car,
  Bike,
  Bus,
  Train,
  Ship,
  Anchor,
  Briefcase,
  Luggage,
  CreditCard,
  Wallet,
  Banknote,
  Coins,
  Receipt,
  Calculator,
  Calendar as CalendarIcon,
  CalendarDays,
  AlarmClock,
  Watch,
  Hourglass,
  StopCircle,
  PlayCircle,
  PauseCircle,
  CheckCircle,
  CheckCircle2,
  XCircle,
  MinusCircle,
  PlusCircle,
  HelpCircle as HelpCircleIcon,
  AlertOctagon,
  Flame,
  Snowflake,
  Leaf,
  TreePine,
  Flower,
  Flower2,
  Bug,
  Rocket,
  Satellite,
  Factory,
  Warehouse,
  Store,
  Hospital,
  Pill,
  Stethoscope,
  Syringe,
  Microscope,
  Dna,
  Bone,
  Brain,
  type LucideIcon,
} from 'lucide-react';

/**
 * Map of string icon names to Lucide React components.
 * This allows dynamic resolution of icons from theme_config.
 */
const iconMap: Record<string, LucideIcon> = {
  // Navigation & Layout
  Home,
  LayoutDashboard,
  MapPin,
  Map,
  Navigation,
  Compass,
  Target,
  Crosshair,
  Globe,
  Layout,
  Layers,
  Grid,
  List,
  Menu,

  // Content & Actions
  ClipboardList,
  Settings,
  Sliders,
  Filter,
  Search,
  Edit,
  Trash2,
  Save,
  Plus,
  Minus,
  Check,
  X,
  Copy,
  Clipboard,
  ClipboardCheck,

  // Users & People
  Users,
  User,
  UserPlus,
  UserMinus,
  UserCog,

  // Buildings & Places
  Building2,
  Hospital,
  Factory,
  Warehouse,
  Store,

  // Time & Calendar
  Calendar: CalendarIcon,
  CalendarDays,
  Clock,
  Timer,
  Alarm: AlarmClock,
  AlarmClock,
  Watch,
  Hourglass,

  // Documents
  FileText,
  File,
  FilePlus,
  FileCheck,
  FileX,
  Folder,
  FolderOpen,

  // History & Status
  History,
  Activity,
  RefreshCw,
  RotateCw,
  RotateCcw,

  // Alerts & Info
  AlertCircle,
  AlertTriangle,
  AlertOctagon,
  Info,
  HelpCircle,
  Bell,

  // Security
  Shield,
  ShieldCheck,
  ShieldAlert,
  Lock,
  Unlock,
  Key,

  // Medical
  Skull,
  Eye,
  EyeOff,
  Heart,
  Pill,
  Stethoscope,
  Syringe,
  Microscope,
  Dna,
  Bone,
  Brain,

  // Communication
  Mail,
  Phone,
  MessageSquare,
  Send,
  Inbox,
  Archive,

  // Media
  Play,
  Pause,
  Stop: Square,
  Square,
  PlayCircle,
  PauseCircle,
  StopCircle,
  Camera,
  Image,
  Video,

  // Charts & Analytics
  BarChart,
  BarChart2,
  PieChart,
  LineChart,
  TrendingUp,
  TrendingDown,

  // Files & Data
  Download,
  Upload,
  Database,
  Server,
  Cloud,
  HardDrive,

  // Arrows & Chevrons
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  ChevronsLeft,
  ChevronsRight,
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ArrowDown,

  // Status indicators
  CheckCircle,
  CheckCircle2,
  XCircle,
  MinusCircle,
  PlusCircle,
  Circle,

  // Misc
  Star,
  Award,
  Gift,
  Bookmark,
  Tag,
  Tags,
  Link,
  Unlink,
  ExternalLink,
  Share,
  Share2,
  Zap,
  Power,
  Flame,
  Rocket,

  // Weather & Environment
  Sun,
  Moon,
  Sunrise,
  Sunset,
  Wind,
  Umbrella,
  Snowflake,
  Thermometer,
  Droplet,
  Leaf,
  TreePine,
  Flower,
  Flower2,

  // Devices
  Monitor,
  Smartphone,
  Tablet,
  Laptop,
  Printer,
  Cpu,
  Wifi,
  WifiOff,
  Bluetooth,
  Usb,

  // Commerce
  ShoppingCart,
  ShoppingBag,
  Package,
  Box,
  CreditCard,
  Wallet,
  Banknote,
  Coins,
  Receipt,
  Calculator,
  DollarSign,
  Percent,

  // Transport
  Truck,
  Plane,
  Car,
  Bike,
  Bus,
  Train,
  Ship,
  Anchor,
  Briefcase,
  Suitcase: Luggage,
  Luggage,

  // Window Controls
  Maximize,
  Minimize,
  Move,
  MoreHorizontal,
  MoreVertical,

  // Audio
  Volume,
  Volume1,
  Volume2,
  VolumeX,
  Mic,
  MicOff,
  Headphones,
  Radio,
  Music,

  // Other
  Cast,
  Tv,
  Film,
  Bug,
  Satellite,
  Hash,
  AtSign,
  Battery,
  BatteryCharging,
  SortAsc,
  SortDesc,
  SkipBack,
  SkipForward,
};

/**
 * Get a Lucide icon component by its string name.
 * Returns Circle icon as fallback for unknown names.
 *
 * @param name - The string name of the icon (e.g., "Home", "Settings")
 * @returns The Lucide icon component
 */
export function getIconByName(name: string): LucideIcon {
  return iconMap[name] || Circle;
}

/**
 * Props for the DynamicIcon component
 */
export interface DynamicIconProps {
  /** The string name of the icon to render */
  name: string;
  /** CSS class name(s) to apply to the icon */
  className?: string;
  /** Icon size in pixels */
  size?: number;
  /** Stroke width for the icon */
  strokeWidth?: number;
}

/**
 * Dynamic icon component that resolves Lucide icons by string name.
 * This allows rendering icons from theme_config JSONB data.
 *
 * @example
 * ```tsx
 * // Render a Home icon
 * <DynamicIcon name="Home" className="h-5 w-5" />
 *
 * // Render with custom size
 * <DynamicIcon name="Settings" size={24} />
 *
 * // Falls back to Circle for unknown names
 * <DynamicIcon name="UnknownIcon" />
 * ```
 */
export function DynamicIcon({
  name,
  className = 'h-5 w-5',
  size,
  strokeWidth,
}: DynamicIconProps) {
  const Icon = getIconByName(name);

  return (
    <Icon
      className={className}
      size={size}
      strokeWidth={strokeWidth}
    />
  );
}

/**
 * List of all available icon names for autocomplete/suggestions
 */
export const availableIconNames = Object.keys(iconMap);

export default DynamicIcon;
