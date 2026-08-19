export type RoverType = "swift" | "hauler";
export type RoverState = "idle" | "moving" | "stranded";
export type ContractStatus =
  | "queued"
  | "accepted"
  | "in_transit"
  | "delivered"
  | "failed"
  | "expired"
  | "lost_to_shadow";

export type HexView = {
  id: string;
  q: number;
  r: number;
  type: string;
  impassable?: boolean;
  inShadow?: boolean;
  phaseEta?: number;
  risk?: string;
};

export type RoverView = {
  type: RoverType;
  name?: string;
  hex: string;
  q: number;
  r: number;
  progress: number;
  battery: number;
  maxBattery: number;
  state: RoverState;
  cargo?: string[];
  path?: string[];
  slotsUsed?: number;
  slotsMax?: number;
  canHeavy?: boolean;
  reversing?: boolean;
  reverseHex?: string;
};

export type Contract = {
  id: string;
  title: string;
  cargoType?: string;
  weight: "light" | "medium" | "heavy";
  pickup: string;
  dropoff: string;
  colonyValue: number;
  earthValue: number;
  reward?: number;
  risk: string;
  urgency: string;
  deadline: number;
  status: ContractStatus;
  assignedTo?: RoverType | "";
  impossible?: boolean;
};

export type Salvage = {
  id: string;
  hex: string;
  value: number;
  status: "available" | "taken" | "lost";
};

export type Snapshot = {
  id?: string;
  seed: string;
  layout?: string;
  status: string;
  outcome?: string;
  endReason?: string;
  t: number;
  rover: RoverView;
  rovers?: RoverView[];
  activeRover: RoverType;
  map: {
    hexes: HexView[];
    terminator: { pos: number; speed: number; direction: "east" | "west" | string };
  };
  contracts: Contract[];
  salvage?: Salvage[];
  crisis: { kind: string; firesAt: number; fired: boolean };
  colonyScore: number;
  earthScore: number;
  autonomyCharges?: number;
  events?: { t: number; kind: string; payload?: Record<string, unknown> }[];
  deltaEvents?: { t: number; kind: string; payload?: Record<string, unknown> }[];
  routePreview?: { feasible: boolean; etaSec?: number; predictedBattery?: number };
  ghost?: { points: { q: number; r: number; t: number }[] };
  goal: { colonyNeed: number; earthSafe: number; duration: number };
  reject?: { reason: string; contractId?: string };
  shareUrl?: string;
  type?: string;
  error?: unknown;
};

export type BlackBox = {
  outcome?: string;
  verdict?: string;
  colonyScore: number;
  earthScore: number;
  seed: string;
  shareUrl?: string;
  endReason?: string;
  events?: { kind: string }[];
};

export type HintKind = "" | "warn" | "ok";
