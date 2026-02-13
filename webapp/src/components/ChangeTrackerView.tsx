import React, { useEffect, useRef, useState, useMemo } from 'react';
import * as d3 from 'd3';
import { ChangeTrackerEventInfo } from '../types/app_types';

export type ChangeTrackerViewProps = {
  data: ChangeTrackerEventInfo[];
  filters: { add: boolean; delete: boolean; change: boolean };
  onPeriodChange: (start: Date, end: Date) => void;
  period: 12 | 6 | 3 | 1; // months
};

const BASE_CELL_SIZE = 16;
const CELL_PADDING_RATIO = 0.125; // 12.5% of cell size for padding
const Y_AXIS_WIDTH_RATIO = 2.5; // Y-axis is 2.5x cell size
const X_AXIS_HEIGHT_RATIO = 1.875; // X-axis is 1.875x cell size
const YAxisWeeks = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

type CellData = {
  date: Date;
  dayOfWeek: number;
  weekIndex: number;
  add: ChangeTrackerEventInfo[];
  change: ChangeTrackerEventInfo[];
  delete: ChangeTrackerEventInfo[];
};

const getEventTypes = (cell: CellData, filters: ChangeTrackerViewProps['filters']) => {
  const types = [];
  if (cell.add.length > 0 && filters.add) types.push('add');
  if (cell.change.length > 0 && filters.change) types.push('change');
  if (cell.delete.length > 0 && filters.delete) types.push('delete');
  return types;
};

const getColorForType = (type: string): string => {
  if (type === 'delete') return '#ef4444';
  if (type === 'add') return '#36a288';
  if (type === 'change') return '#eab308';
  return '#e5e7eb';
};

const ChangeTrackerView = (props: ChangeTrackerViewProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const [containerWidth, setContainerWidth] = useState(1000);
  const [tooltip, setTooltip] = useState<{
    visible: boolean;
    x: number;
    y: number;
    content: string;
  }>({ visible: false, x: 0, y: 0, content: '' });

  // Generate all dates in the range and group events
  const cellData = useMemo(() => {
    const endDate = new Date();
    endDate.setHours(0, 0, 0, 0);

    const startDate = new Date();
    startDate.setMonth(startDate.getMonth() - props.period);
    startDate.setDate(startDate.getDate() - startDate.getDay()); // Start from Sunday
    startDate.setHours(0, 0, 0, 0);

    const grouped: { [key: string]: CellData } = {};

    // Initialize all dates in range
    const currentDate = new Date(startDate);
    while (currentDate <= endDate) {
      const key = currentDate.toISOString().split('T')[0];
      const weekIndex = Math.floor(
        (currentDate.getTime() - startDate.getTime()) / (7 * 24 * 60 * 60 * 1000)
      );

      grouped[key] = {
        date: new Date(currentDate),
        dayOfWeek: currentDate.getDay(),
        weekIndex,
        add: [],
        change: [],
        delete: [],
      };

      currentDate.setDate(currentDate.getDate() + 1);
    }

    // Populate with actual events
    props.data.forEach((event) => {
      const eventDate = new Date(event.timestamp);
      eventDate.setHours(0, 0, 0, 0);
      const key = eventDate.toISOString().split('T')[0];

      if (grouped[key]) {
        if (event.type === 'add') {
          grouped[key].add.push(event);
        } else if (event.type === 'change') {
          grouped[key].change.push(event);
        } else if (event.type === 'delete') {
          grouped[key].delete.push(event);
        }
      }
    });

    const result = Object.values(grouped);

    // Call the callback with date range
    props.onPeriodChange(startDate, endDate);

    return result;
    // Data visualization is not yet complete. So We cannot verify this suggestions for now.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.data, props.period, props.onPeriodChange]);

  // Calculate max week index
  const maxWeekIndex = useMemo(() => {
    return Math.max(...cellData.map((c) => c.weekIndex), 0);
  }, [cellData]);

  // Calculate scaled dimensions based on container width
  const scaledDimensions = useMemo(() => {
    const numWeeks = maxWeekIndex + 1;
    const numDays = 7;

    // Calculate what cell size would fit the container width
    const availableWidth = containerWidth - 40; // Account for padding
    const yAxisWidth = BASE_CELL_SIZE * Y_AXIS_WIDTH_RATIO;
    const widthForGrid = availableWidth - yAxisWidth;

    // Calculate cell size to fit width
    const cellSize = Math.floor(widthForGrid / (numWeeks + (numWeeks - 1) * CELL_PADDING_RATIO));
    const actualCellSize = Math.max(12, Math.min(cellSize, 24)); // Min 12px, max 24px

    const cellPadding = actualCellSize * CELL_PADDING_RATIO;
    const scaledYAxisWidth = actualCellSize * Y_AXIS_WIDTH_RATIO;
    const scaledXAxisHeight = actualCellSize * X_AXIS_HEIGHT_RATIO;

    const gridWidth = numWeeks * (actualCellSize + cellPadding) - cellPadding;
    const gridHeight = numDays * (actualCellSize + cellPadding) - cellPadding;

    const svgWidth = gridWidth + scaledYAxisWidth + 20;
    const svgHeight = gridHeight + scaledXAxisHeight + 40;

    return {
      cellSize: actualCellSize,
      cellPadding,
      yAxisWidth: scaledYAxisWidth,
      xAxisHeight: scaledXAxisHeight,
      svgWidth,
      svgHeight,
      fontSize: Math.max(10, actualCellSize * 0.6875), // Scale font size proportionally
    };
  }, [containerWidth, maxWeekIndex]);

  // Update container width on resize
  useEffect(() => {
    const updateWidth = () => {
      if (containerRef.current) {
        const { width } = containerRef.current.getBoundingClientRect();
        setContainerWidth(width);
      }
    };

    updateWidth();
    window.addEventListener('resize', updateWidth);

    const resizeObserver = new ResizeObserver(updateWidth);
    if (containerRef.current) {
      resizeObserver.observe(containerRef.current);
    }

    return () => {
      window.removeEventListener('resize', updateWidth);
      resizeObserver.disconnect();
    };
  }, []);

  useEffect(() => {
    if (!svgRef.current || cellData.length === 0) return;

    const { cellSize, cellPadding, yAxisWidth, xAxisHeight, fontSize } = scaledDimensions;
    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    // Add month labels (X-axis)
    const monthMap = new Map<string, number>();
    cellData.forEach((d) => {
      const key = `${d.date.getFullYear()}-${d.date.getMonth()}`;
      if (!monthMap.has(key)) {
        monthMap.set(key, d.date.getMonth());
      }
    });

    const monthLabelsGroup = svg.append('g').attr('class', 'month-labels');

    const monthLabelData: Array<{ month: string; weekIndex: number }> = [];
    monthMap.forEach((monthNum, key) => {
      const cells = cellData.filter((d) => `${d.date.getFullYear()}-${d.date.getMonth()}` === key);
      if (cells.length > 0) {
        const selectedCell = cells.length === 1 ? cells[0] : cells[Math.floor(cells.length / 2)];

        monthLabelData.push({
          month: new Date(selectedCell.date.getFullYear(), selectedCell.date.getMonth())
            .toLocaleDateString('en-US', { month: 'short' })
            .toUpperCase(),
          weekIndex: selectedCell.weekIndex,
        });
      }
    });

    monthLabelsGroup
      .selectAll('text')
      .data(monthLabelData)
      .enter()
      .append('text')
      .attr('x', (d) => yAxisWidth + d.weekIndex * (cellSize + cellPadding))
      .attr('y', 15)
      .attr('font-size', `${fontSize}px`)
      .attr('fill', '#666')
      .attr('text-anchor', 'start')
      .text((d) => d.month);

    // Add day labels (Y-axis)
    const dayLabelsGroup = svg.append('g').attr('class', 'day-labels');
    dayLabelsGroup
      .selectAll('text')
      .data(YAxisWeeks)
      .enter()
      .append('text')
      .attr('x', yAxisWidth - 10)
      .attr('y', (d, i) => xAxisHeight + i * (cellSize + cellPadding) + cellSize - 2)
      .attr('font-size', `${fontSize}px`)
      .attr('fill', '#666')
      .attr('text-anchor', 'end')
      .text((d) => d);

    // Add cells
    const cellsGroup = svg.append('g').attr('class', 'cells');

    cellData.forEach((d) => {
      const x = yAxisWidth + d.weekIndex * (cellSize + cellPadding);
      const y = xAxisHeight + d.dayOfWeek * (cellSize + cellPadding);
      const eventTypes = getEventTypes(d, props.filters);

      const cellGroup = cellsGroup
        .append('g')
        .attr('class', 'cell-group')
        .style('cursor', 'pointer');

      const borderRadius = Math.max(2, cellSize * 0.125);

      if (eventTypes.length === 0) {
        // No activity - single gray cell
        cellGroup
          .append('rect')
          .attr('x', x)
          .attr('y', y)
          .attr('width', cellSize)
          .attr('height', cellSize)
          .attr('fill', '#e5e7eb')
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);
      } else if (eventTypes.length === 1) {
        // Single event type - single colored cell
        cellGroup
          .append('rect')
          .attr('x', x)
          .attr('y', y)
          .attr('width', cellSize)
          .attr('height', cellSize)
          .attr('fill', getColorForType(eventTypes[0]))
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);
      } else if (eventTypes.length === 2) {
        // Two event types - split horizontally
        cellGroup
          .append('rect')
          .attr('x', x)
          .attr('y', y)
          .attr('width', cellSize)
          .attr('height', cellSize / 2)
          .attr('fill', getColorForType(eventTypes[0]))
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);

        cellGroup
          .append('rect')
          .attr('x', x)
          .attr('y', y + cellSize / 2)
          .attr('width', cellSize)
          .attr('height', cellSize / 2)
          .attr('fill', getColorForType(eventTypes[1]))
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);
      } else if (eventTypes.length === 3) {
        // Three event types - split as per design
        cellGroup
          .append('rect')
          .attr('x', x)
          .attr('y', y)
          .attr('width', cellSize / 2)
          .attr('height', cellSize / 2)
          .attr('fill', getColorForType('add'))
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);

        cellGroup
          .append('rect')
          .attr('x', x + cellSize / 2)
          .attr('y', y)
          .attr('width', cellSize / 2)
          .attr('height', cellSize / 2)
          .attr('fill', getColorForType('change'))
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);

        cellGroup
          .append('rect')
          .attr('x', x)
          .attr('y', y + cellSize / 2)
          .attr('width', cellSize)
          .attr('height', cellSize / 2)
          .attr('fill', getColorForType('delete'))
          .attr('stroke', '#e5e7eb')
          .attr('stroke-width', 0.5)
          .attr('rx', borderRadius);
      }

      // Add invisible overlay for hover/click events
      const overlay = cellGroup
        .append('rect')
        .attr('x', x)
        .attr('y', y)
        .attr('width', cellSize)
        .attr('height', cellSize)
        .attr('fill', 'transparent')
        .attr('stroke', '#e5e7eb')
        .attr('stroke-width', 0.5)
        .attr('rx', borderRadius);

      cellGroup
        .on('mouseenter', function () {
          overlay.attr('stroke', '#14b8a6').attr('stroke-width', 2);

          const dateStr = d.date.toLocaleDateString('en-US', {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
          });

          const contentParts = [dateStr, ''];

          if (d.add.length > 0) {
            contentParts.push(`Added (${d.add.length}):`);
            d.add.forEach((event) => {
              contentParts.push(`  • ${event.message || 'No message'}`);
            });
          }

          if (d.change.length > 0) {
            if (contentParts.length > 2) contentParts.push('');
            contentParts.push(`Changed (${d.change.length}):`);
            d.change.forEach((event) => {
              contentParts.push(`  • ${event.message || 'No message'}`);
            });
          }

          if (d.delete.length > 0) {
            if (contentParts.length > 2) contentParts.push('');
            contentParts.push(`Deleted (${d.delete.length}):`);
            d.delete.forEach((event) => {
              contentParts.push(`  • ${event.message || 'No message'}`);
            });
          }

          if (d.add.length === 0 && d.change.length === 0 && d.delete.length === 0) {
            contentParts.push('No activity');
          }

          const rect = overlay.node()?.getBoundingClientRect();
          if (rect) {
            setTooltip({
              visible: true,
              x: Math.round(rect.left + rect.width / 2),
              y: Math.round(rect.top - 10),
              content: contentParts.join('\n'),
            });
          }
        })
        .on('mouseleave', function () {
          overlay.attr('stroke', '#e5e7eb').attr('stroke-width', 0.5);
          setTooltip({ visible: false, x: 0, y: 0, content: '' });
        });
    });
  }, [cellData, props.filters, scaledDimensions]);

  return (
    <div ref={containerRef} className="w-full relative">
      <div className="overflow-x-auto overflow-y-hidden border-round-lg bg-white p-2 pb-0">
        <svg
          ref={svgRef}
          width={scaledDimensions.svgWidth}
          height={scaledDimensions.svgHeight}
          className="block"
          style={{
            minWidth: '100%',
          }}
        />
      </div>
      {tooltip.visible && (
        <div
          className="fixed bg-gray-900 text-white text-xs px-3 py-2 border-round shadow-3 pointer-events-none z-5 white-space-pre-line"
          style={{
            left: `${tooltip.x}px`,
            top: `${tooltip.y}px`,
            transform: 'translate(-50%, -100%)',
            maxWidth: '400px',
            maxHeight: '400px',
            overflowY: 'auto',
          }}
        >
          {tooltip.content}
        </div>
      )}
    </div>
  );
};

export default ChangeTrackerView;
