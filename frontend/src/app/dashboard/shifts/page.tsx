'use client';

import './calendar.css';
import { useState, useEffect, useRef, useMemo } from 'react';
import FullCalendar from '@fullcalendar/react';
import dayGridPlugin from '@fullcalendar/daygrid';
import timeGridPlugin from '@fullcalendar/timegrid';
import interactionPlugin from '@fullcalendar/interaction';
import listPlugin from '@fullcalendar/list';
import type { EventInput, EventClickArg, DateSelectArg } from '@fullcalendar/core';
import ptBrLocale from '@fullcalendar/core/locales/pt-br';
import {
  Calendar,
  Clock,
  Users,
  AlertTriangle,
  Plus,
  Loader2,
  Sun,
  Moon,
  ChevronLeft,
  ChevronRight,
  List,
  CalendarDays,
  LayoutGrid,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  useHospitals,
  useShifts,
  useTodayShifts,
  useCoverageGaps,
  useCreateShift,
  useUpdateShift,
  useDeleteShift,
} from '@/hooks';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import type { Shift, DayOfWeek, User } from '@/types';
import { DayNames } from '@/types';

interface ShiftFormData {
  user_id: string;
  day_of_week: DayOfWeek;
  start_time: string;
  end_time: string;
}

// Color palette for users (assigned dynamically)
const userColors = [
  { bg: '#3B82F6', border: '#2563EB', text: '#FFFFFF' }, // Blue
  { bg: '#10B981', border: '#059669', text: '#FFFFFF' }, // Green
  { bg: '#8B5CF6', border: '#7C3AED', text: '#FFFFFF' }, // Purple
  { bg: '#F59E0B', border: '#D97706', text: '#FFFFFF' }, // Amber
  { bg: '#EF4444', border: '#DC2626', text: '#FFFFFF' }, // Red
  { bg: '#EC4899', border: '#DB2777', text: '#FFFFFF' }, // Pink
  { bg: '#06B6D4', border: '#0891B2', text: '#FFFFFF' }, // Cyan
  { bg: '#84CC16', border: '#65A30D', text: '#FFFFFF' }, // Lime
  { bg: '#F97316', border: '#EA580C', text: '#FFFFFF' }, // Orange
  { bg: '#6366F1', border: '#4F46E5', text: '#FFFFFF' }, // Indigo
];

function getUserColor(userId: string, userColorMap: Map<string, typeof userColors[0]>) {
  if (!userColorMap.has(userId)) {
    const colorIndex = userColorMap.size % userColors.length;
    userColorMap.set(userId, userColors[colorIndex]);
  }
  return userColorMap.get(userId)!;
}

// Convert day_of_week (0=Sunday) to next occurrence of that day
function getNextDayOfWeek(dayOfWeek: number): Date {
  const today = new Date();
  const currentDay = today.getDay();
  const diff = dayOfWeek - currentDay;
  const nextDate = new Date(today);
  nextDate.setDate(today.getDate() + diff);
  return nextDate;
}

// Generate recurring events for each shift (for the visible calendar range)
function shiftsToEvents(shifts: Shift[], userColorMap: Map<string, typeof userColors[0]>): EventInput[] {
  const events: EventInput[] = [];
  const today = new Date();

  // Generate events for 8 weeks (past 4 + future 4)
  for (let weekOffset = -4; weekOffset <= 4; weekOffset++) {
    shifts.forEach((shift) => {
      const baseDate = new Date(today);
      baseDate.setDate(today.getDate() + (weekOffset * 7));

      const eventDate = new Date(baseDate);
      const currentDay = eventDate.getDay();
      const diff = shift.day_of_week - currentDay;
      eventDate.setDate(eventDate.getDate() + diff);

      const [startHour, startMin] = shift.start_time.split(':').map(Number);
      const [endHour, endMin] = shift.end_time.split(':').map(Number);

      const start = new Date(eventDate);
      start.setHours(startHour, startMin, 0, 0);

      const end = new Date(eventDate);
      end.setHours(endHour, endMin, 0, 0);

      // Handle overnight shifts
      if (endHour < startHour || (endHour === startHour && endMin < startMin)) {
        end.setDate(end.getDate() + 1);
      }

      const color = getUserColor(shift.user_id, userColorMap);

      events.push({
        id: `${shift.id}-${weekOffset}`,
        title: shift.user?.nome || 'Operador',
        start,
        end,
        backgroundColor: shift.is_night ? '#4F46E5' : color.bg,
        borderColor: shift.is_night ? '#3730A3' : color.border,
        textColor: color.text,
        extendedProps: {
          shiftId: shift.id,
          shift,
          isNight: shift.is_night,
          userName: shift.user?.nome || 'Operador',
        },
      });
    });
  }

  return events;
}

function ShiftFormDialog({
  open,
  onOpenChange,
  hospitalId,
  shift,
  users,
  onSubmit,
  onDelete,
  isLoading,
  initialDate,
  initialTime,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  hospitalId: string;
  shift?: Shift | null;
  users: User[];
  onSubmit: (data: ShiftFormData) => void;
  onDelete?: () => void;
  isLoading: boolean;
  initialDate?: Date;
  initialTime?: { start: string; end: string };
}) {
  const [formData, setFormData] = useState<ShiftFormData>({
    user_id: '',
    day_of_week: 1 as DayOfWeek,
    start_time: '08:00',
    end_time: '18:00',
  });

  useEffect(() => {
    if (shift) {
      setFormData({
        user_id: shift.user_id,
        day_of_week: shift.day_of_week,
        start_time: shift.start_time,
        end_time: shift.end_time,
      });
    } else if (initialDate) {
      setFormData({
        user_id: '',
        day_of_week: initialDate.getDay() as DayOfWeek,
        start_time: initialTime?.start || '08:00',
        end_time: initialTime?.end || '18:00',
      });
    } else {
      setFormData({
        user_id: '',
        day_of_week: 1 as DayOfWeek,
        start_time: '08:00',
        end_time: '18:00',
      });
    }
  }, [shift, open, initialDate, initialTime]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {shift ? (
              <>
                <Clock className="h-5 w-5" />
                Editar Escala
              </>
            ) : (
              <>
                <Plus className="h-5 w-5" />
                Nova Escala
              </>
            )}
          </DialogTitle>
          <DialogDescription>
            {shift
              ? 'Altere os dados da escala de plantao.'
              : 'Cadastre uma nova escala de plantao.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="user_id">Operador</Label>
              <Select
                value={formData.user_id}
                onValueChange={(value) => setFormData({ ...formData, user_id: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione o operador" />
                </SelectTrigger>
                <SelectContent>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      <div className="flex items-center gap-2">
                        <div className="h-2 w-2 rounded-full bg-primary" />
                        {user.nome}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="day_of_week">Dia da Semana</Label>
              <Select
                value={formData.day_of_week.toString()}
                onValueChange={(value) =>
                  setFormData({ ...formData, day_of_week: parseInt(value) as DayOfWeek })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione o dia" />
                </SelectTrigger>
                <SelectContent>
                  {([0, 1, 2, 3, 4, 5, 6] as DayOfWeek[]).map((day) => (
                    <SelectItem key={day} value={day.toString()}>
                      {DayNames[day]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="start_time" className="flex items-center gap-1">
                  <Sun className="h-3 w-3" />
                  Inicio
                </Label>
                <Input
                  id="start_time"
                  type="time"
                  value={formData.start_time}
                  onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="end_time" className="flex items-center gap-1">
                  <Moon className="h-3 w-3" />
                  Fim
                </Label>
                <Input
                  id="end_time"
                  type="time"
                  value={formData.end_time}
                  onChange={(e) => setFormData({ ...formData, end_time: e.target.value })}
                  required
                />
              </div>
            </div>
          </div>
          <DialogFooter className="flex justify-between sm:justify-between">
            {shift && onDelete && (
              <Button type="button" variant="destructive" onClick={onDelete}>
                Excluir
              </Button>
            )}
            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                Cancelar
              </Button>
              <Button type="submit" disabled={isLoading || !formData.user_id}>
                {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {shift ? 'Salvar' : 'Criar'}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function CoverageAlert({ hospitalId }: { hospitalId: string }) {
  const { data: coverage } = useCoverageGaps(hospitalId);

  if (!coverage?.has_gaps) return null;

  return (
    <div className="flex items-center gap-2 px-3 py-2 bg-orange-100 dark:bg-orange-900/30 border border-orange-200 dark:border-orange-800 rounded-lg text-sm">
      <AlertTriangle className="h-4 w-4 text-orange-600 dark:text-orange-400" />
      <span className="text-orange-800 dark:text-orange-200">
        {coverage.gaps.length} periodo(s) sem cobertura
      </span>
    </div>
  );
}

function TodayShiftsBadge({ hospitalId }: { hospitalId: string }) {
  const { data: todayShifts = [] } = useTodayShifts(hospitalId);
  const activeShift = todayShifts.find((s) => s.is_active);

  if (!activeShift) return null;

  return (
    <Badge variant="default" className="gap-1">
      <Clock className="h-3 w-3" />
      {activeShift.user?.nome} em plantao
    </Badge>
  );
}

export default function ShiftsPage() {
  const { user } = useAuth();
  const { data: hospitals = [] } = useHospitals();
  const calendarRef = useRef<FullCalendar>(null);

  const [selectedHospitalId, setSelectedHospitalId] = useState<string>('');
  const [users, setUsers] = useState<User[]>([]);
  const [currentView, setCurrentView] = useState<'timeGridWeek' | 'dayGridMonth' | 'listWeek'>('timeGridWeek');
  const [calendarTitle, setCalendarTitle] = useState('');

  // Dialog states
  const [formDialogOpen, setFormDialogOpen] = useState(false);
  const [editingShift, setEditingShift] = useState<Shift | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deletingShift, setDeletingShift] = useState<Shift | null>(null);
  const [initialDate, setInitialDate] = useState<Date | undefined>();
  const [initialTime, setInitialTime] = useState<{ start: string; end: string } | undefined>();

  // User color map for consistent coloring
  const userColorMap = useMemo(() => new Map<string, typeof userColors[0]>(), []);

  const effectiveHospitalId = selectedHospitalId || (user?.hospital_id ?? '');

  const { data: shifts = [], isLoading: shiftsLoading } = useShifts(effectiveHospitalId);

  const createShift = useCreateShift();
  const updateShift = useUpdateShift();
  const deleteShift = useDeleteShift();

  const canManage = user?.role === 'admin' || user?.role === 'gestor';
  const canSelectHospital = user?.role === 'admin';

  // Fetch users
  useEffect(() => {
    async function fetchUsers() {
      if (!effectiveHospitalId) {
        setUsers([]);
        return;
      }
      try {
        const response = await api.get<User[] | { data: User[] }>('/users');
        const userData = Array.isArray(response.data) ? response.data : response.data.data;
        setUsers(userData || []);
      } catch (error) {
        console.error('Error fetching users:', error);
        setUsers([]);
      }
    }
    fetchUsers();
  }, [effectiveHospitalId]);

  // Convert shifts to calendar events
  const events = useMemo(() => {
    return shiftsToEvents(shifts, userColorMap);
  }, [shifts, userColorMap]);

  // Calendar navigation
  const handlePrev = () => calendarRef.current?.getApi().prev();
  const handleNext = () => calendarRef.current?.getApi().next();
  const handleToday = () => calendarRef.current?.getApi().today();

  const handleViewChange = (view: typeof currentView) => {
    setCurrentView(view);
    calendarRef.current?.getApi().changeView(view);
  };

  // Handle date/time selection on calendar
  const handleDateSelect = (selectInfo: DateSelectArg) => {
    if (!canManage) return;

    const startDate = selectInfo.start;
    const endDate = selectInfo.end;

    setInitialDate(startDate);
    setInitialTime({
      start: startDate.toTimeString().slice(0, 5),
      end: endDate.toTimeString().slice(0, 5),
    });
    setEditingShift(null);
    setFormDialogOpen(true);

    calendarRef.current?.getApi().unselect();
  };

  // Handle click on existing event
  const handleEventClick = (clickInfo: EventClickArg) => {
    if (!canManage) return;

    const shift = clickInfo.event.extendedProps.shift as Shift;
    setEditingShift(shift);
    setInitialDate(undefined);
    setInitialTime(undefined);
    setFormDialogOpen(true);
  };

  const handleFormSubmit = async (data: ShiftFormData) => {
    try {
      if (editingShift) {
        await updateShift.mutateAsync({
          id: editingShift.id,
          ...data,
        });
      } else {
        await createShift.mutateAsync({
          hospital_id: effectiveHospitalId,
          ...data,
        });
      }
      setFormDialogOpen(false);
      setEditingShift(null);
    } catch (error) {
      console.error('Error saving shift:', error);
    }
  };

  const handleDeleteClick = () => {
    if (editingShift) {
      setDeletingShift(editingShift);
      setDeleteDialogOpen(true);
    }
  };

  const handleConfirmDelete = async () => {
    if (!deletingShift) return;
    try {
      await deleteShift.mutateAsync(deletingShift.id);
      setDeleteDialogOpen(false);
      setDeletingShift(null);
      setFormDialogOpen(false);
      setEditingShift(null);
    } catch (error) {
      console.error('Error deleting shift:', error);
    }
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <Calendar className="h-6 w-6" />
            Escalas de Plantao
          </h1>
          <p className="text-muted-foreground">
            {canManage ? 'Clique e arraste para criar escalas' : 'Visualize as escalas dos operadores'}
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          {effectiveHospitalId && <TodayShiftsBadge hospitalId={effectiveHospitalId} />}
          {effectiveHospitalId && <CoverageAlert hospitalId={effectiveHospitalId} />}

          {canSelectHospital && (
            <Select value={selectedHospitalId} onValueChange={setSelectedHospitalId}>
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="Selecione hospital" />
              </SelectTrigger>
              <SelectContent>
                {hospitals.map((hospital) => (
                  <SelectItem key={hospital.id} value={hospital.id}>
                    {hospital.nome}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {canManage && effectiveHospitalId && (
            <Button onClick={() => {
              setEditingShift(null);
              setInitialDate(undefined);
              setInitialTime(undefined);
              setFormDialogOpen(true);
            }}>
              <Plus className="mr-2 h-4 w-4" />
              Nova Escala
            </Button>
          )}
        </div>
      </div>

      {!effectiveHospitalId ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Calendar className="h-16 w-16 text-muted-foreground mb-4" />
            <p className="text-lg text-muted-foreground">Selecione um hospital para ver as escalas</p>
          </CardContent>
        </Card>
      ) : (
        <Card className="overflow-hidden">
          {/* Calendar Toolbar */}
          <div className="flex items-center justify-between p-3 border-b bg-muted/30">
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={handleToday}>
                Hoje
              </Button>
              <div className="flex items-center">
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handlePrev}>
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleNext}>
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
              <span className="font-semibold text-lg min-w-[200px]">{calendarTitle}</span>
            </div>

            <TooltipProvider>
              <div className="flex items-center gap-1 bg-muted rounded-lg p-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={currentView === 'timeGridWeek' ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-8"
                      onClick={() => handleViewChange('timeGridWeek')}
                    >
                      <CalendarDays className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Semana</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={currentView === 'dayGridMonth' ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-8"
                      onClick={() => handleViewChange('dayGridMonth')}
                    >
                      <LayoutGrid className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Mes</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={currentView === 'listWeek' ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-8"
                      onClick={() => handleViewChange('listWeek')}
                    >
                      <List className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Lista</TooltipContent>
                </Tooltip>
              </div>
            </TooltipProvider>
          </div>

          {/* Calendar */}
          <div className="p-2">
            {shiftsLoading ? (
              <div className="flex items-center justify-center h-[600px]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <FullCalendar
                ref={calendarRef}
                plugins={[dayGridPlugin, timeGridPlugin, interactionPlugin, listPlugin]}
                initialView={currentView}
                locale={ptBrLocale}
                headerToolbar={false}
                events={events}
                selectable={canManage}
                selectMirror={true}
                dayMaxEvents={true}
                weekends={true}
                nowIndicator={true}
                allDaySlot={false}
                slotMinTime="00:00:00"
                slotMaxTime="24:00:00"
                slotDuration="01:00:00"
                slotLabelInterval="02:00:00"
                height={600}
                select={handleDateSelect}
                eventClick={handleEventClick}
                datesSet={(dateInfo) => {
                  const formatter = new Intl.DateTimeFormat('pt-BR', {
                    month: 'long',
                    year: 'numeric',
                  });
                  setCalendarTitle(formatter.format(dateInfo.start).replace(/^\w/, c => c.toUpperCase()));
                }}
                eventContent={(eventInfo) => {
                  const isNight = eventInfo.event.extendedProps.isNight;
                  return (
                    <div className="flex items-center gap-1 px-1 py-0.5 overflow-hidden">
                      {isNight ? (
                        <Moon className="h-3 w-3 shrink-0" />
                      ) : (
                        <Sun className="h-3 w-3 shrink-0" />
                      )}
                      <span className="truncate text-xs font-medium">
                        {eventInfo.event.title}
                      </span>
                    </div>
                  );
                }}
                eventClassNames="cursor-pointer rounded-md shadow-sm"
              />
            )}
          </div>

          {/* Legend */}
          <div className="flex items-center gap-4 p-3 border-t bg-muted/30 text-sm">
            <div className="flex items-center gap-2">
              <Sun className="h-4 w-4 text-yellow-600" />
              <span className="text-muted-foreground">Diurno</span>
            </div>
            <div className="flex items-center gap-2">
              <Moon className="h-4 w-4 text-indigo-600" />
              <span className="text-muted-foreground">Noturno</span>
            </div>
            <div className="ml-auto text-muted-foreground">
              {shifts.length} escala(s) cadastrada(s)
            </div>
          </div>
        </Card>
      )}

      {/* Create/Edit Dialog */}
      <ShiftFormDialog
        open={formDialogOpen}
        onOpenChange={setFormDialogOpen}
        hospitalId={effectiveHospitalId}
        shift={editingShift}
        users={users}
        onSubmit={handleFormSubmit}
        onDelete={editingShift ? handleDeleteClick : undefined}
        isLoading={createShift.isPending || updateShift.isPending}
        initialDate={initialDate}
        initialTime={initialTime}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirmar exclusao</AlertDialogTitle>
            <AlertDialogDescription>
              Tem certeza que deseja excluir a escala de{' '}
              <strong>{deletingShift?.user?.nome}</strong> no dia{' '}
              <strong>{deletingShift ? DayNames[deletingShift.day_of_week] : ''}</strong>?
              <br />
              Esta acao nao pode ser desfeita.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteShift.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Excluir
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
