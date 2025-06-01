// Copyright (C) 2025 Aaron Mathis
// This file is part of GoSight Server.
//
// GoSight Server is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoSight Server is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with GoSight Server.  If not, see <https://www.gnu.org/licenses/>.

/**
 * Dashboard types for the GoSight dashboard system.
 * Defines the structure for widgets, dashboards, and their positioning.
 */

export interface WidgetPosition {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface Widget {
  id: string;
  type: string;
  title: string;
  position: WidgetPosition;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface Dashboard {
  id: string;
  name: string;
  widgets: Widget[];
  layout: {
    columns: number;
    rowHeight: number;
  };
  createdAt: string;
  updatedAt: string;
}

export type WidgetType = 
  | 'metric-card'
  | 'chart'
  | 'table'
  | 'log-viewer'
  | 'alert-list'
  | 'status-indicator';

export interface WidgetTemplate {
  type: WidgetType;
  name: string;
  description: string;
  defaultSize: Pick<WidgetPosition, 'width' | 'height'>;
  icon: string;
}
