# Task Breakdown: Cadastro de Hospitais com Integracao ao Mapa

## Overview
Total Tasks: 32
Estimated Total Complexity: Medium-High

## Task List

### Backend Layer

#### Task Group 1: Database Migration and Model Updates
**Dependencies:** None
**Complexity:** Low
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/backend/migrations/018_add_telefone_to_hospitals.sql` (new)
- `/home/matheus_rubem/VitalConnect/backend/internal/models/hospital.go`

- [x] 1.0 Complete database layer for telefone field
  - [x] 1.1 Write 3-4 focused tests for Hospital model with telefone
    - Test telefone field persistence (create/read)
    - Test telefone optional validation (null allowed)
    - Test telefone format validation if populated
    - Test model ToResponse includes telefone
  - [x] 1.2 Create migration 018_add_telefone_to_hospitals.sql
    - Add telefone column VARCHAR(20) nullable
    - Add comment describing format: (XX) XXXX-XXXX or (XX) XXXXX-XXXX
    - Follow existing migration pattern from 017_add_coordinates_to_hospitals.sql
  - [x] 1.3 Update Hospital model struct
    - Add Telefone field: `Telefone *string json:"telefone,omitempty" db:"telefone"`
    - Add validation tag: `validate:"omitempty,max=20"`
  - [x] 1.4 Update CreateHospitalInput struct
    - Add Telefone field with optional validation
  - [x] 1.5 Update UpdateHospitalInput struct
    - Add Telefone field with optional validation
  - [x] 1.6 Update HospitalResponse struct
    - Add Telefone field for API response
  - [x] 1.7 Update ToResponse method
    - Include Telefone in response mapping
  - [x] 1.8 Ensure database layer tests pass
    - Run ONLY the 3-4 tests written in 1.1
    - Verify migration runs successfully

**Acceptance Criteria:**
- Migration 018 creates telefone column successfully
- Hospital model includes telefone field with proper validation
- CreateHospitalInput and UpdateHospitalInput support telefone
- HospitalResponse includes telefone in API output
- Tests from 1.1 pass

---

#### Task Group 2: Backend Repository Updates
**Dependencies:** Task Group 1
**Complexity:** Low
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/backend/internal/repository/hospital_repository.go`

- [x] 2.0 Complete repository layer updates
  - [x] 2.1 Write 2-3 focused tests for repository telefone handling
    - Test Create with telefone populated
    - Test Update with telefone field
    - Test List returns telefone in results
  - [x] 2.2 Update Create method SQL query
    - Include telefone in INSERT statement
    - Include telefone in RETURNING clause
  - [x] 2.3 Update Update method SQL query
    - Handle telefone in UPDATE statement (when provided)
  - [x] 2.4 Verify List and GetByID return telefone
    - Ensure SELECT queries include telefone column
  - [x] 2.5 Ensure repository tests pass
    - Run ONLY the 2-3 tests written in 2.1

**Acceptance Criteria:**
- Create operation persists telefone correctly
- Update operation modifies telefone when provided
- All read operations return telefone field
- Tests from 2.1 pass

---

#### Task Group 3: Backend Validation Enhancement
**Dependencies:** Task Group 2
**Complexity:** Low
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/backend/internal/handlers/hospitals.go`
- `/home/matheus_rubem/VitalConnect/backend/internal/models/hospital.go` (validation rules)

- [x] 3.0 Complete backend validation for coordinates requirement
  - [x] 3.1 Write 3-4 focused tests for API validation
    - Test POST without coordinates returns 400
    - Test POST with valid coordinates returns 201
    - Test PATCH allows partial updates (coordinates not required)
    - Test telefone format validation if provided
  - [x] 3.2 Update CreateHospitalInput validation
    - Make Latitude required: `validate:"required,min=-90,max=90"`
    - Make Longitude required: `validate:"required,min=-180,max=180"`
    - Keep Endereco required
  - [x] 3.3 Keep UpdateHospitalInput flexible
    - Coordinates remain optional for PATCH (partial updates)
  - [x] 3.4 Add telefone format validation (optional)
    - Custom validator for Brazilian phone format if populated
    - Pattern: (XX) XXXX-XXXX or (XX) XXXXX-XXXX
  - [x] 3.5 Ensure validation tests pass
    - Run ONLY the 3-4 tests written in 3.1

**Acceptance Criteria:**
- POST /hospitals requires latitude and longitude
- POST /hospitals requires endereco
- PATCH /hospitals/:id allows partial updates
- Telefone validates Brazilian format when provided
- Tests from 3.1 pass

---

### Frontend Layer

#### Task Group 4: Frontend Types and API Service
**Dependencies:** Task Group 3
**Complexity:** Low
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/frontend/src/types/index.ts`
- `/home/matheus_rubem/VitalConnect/frontend/src/hooks/useHospitals.ts`

- [x] 4.0 Complete frontend types and API integration
  - [x] 4.1 Write 2-3 focused tests for hospital mutations
    - Test createHospital mutation calls correct endpoint
    - Test updateHospital mutation calls correct endpoint
    - Test mutation invalidates hospitals query cache
  - [x] 4.2 Update Hospital interface in types/index.ts
    - Add telefone?: string field
    - Add latitude?: number field
    - Add longitude?: number field
  - [x] 4.3 Create CreateHospitalInput interface
    - nome: string (required)
    - codigo: string (required)
    - endereco: string (required)
    - telefone?: string (optional)
    - latitude: number (required)
    - longitude: number (required)
    - ativo: boolean (default true)
  - [x] 4.4 Create UpdateHospitalInput interface
    - All fields optional for partial updates
  - [x] 4.5 Extend useHospitals hook with mutations
    - Add createHospital mutation using useMutation
    - Add updateHospital mutation using useMutation
    - Invalidate queryKey ['hospitals'] on success
    - Follow pattern from existing hooks
  - [x] 4.6 Ensure frontend API tests pass
    - Run ONLY the 2-3 tests written in 4.1

**Acceptance Criteria:**
- Hospital type includes telefone, latitude, longitude
- CreateHospitalInput and UpdateHospitalInput defined
- useHospitals hook exposes createHospital and updateHospital mutations
- Cache invalidation works correctly
- Tests from 4.1 pass

---

#### Task Group 5: Geocoding Service
**Dependencies:** None (can run in parallel with Task Groups 1-4)
**Complexity:** Medium
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/frontend/src/services/nominatim.ts` (new)
- `/home/matheus_rubem/VitalConnect/frontend/src/hooks/useGeocoding.ts` (new)

- [x] 5.0 Complete Nominatim geocoding integration
  - [x] 5.1 Write 3-4 focused tests for geocoding service
    - Test searchAddress returns array of suggestions
    - Test debounce prevents excessive API calls
    - Test empty query returns empty array
    - Test error handling for API failures
  - [x] 5.2 Create nominatim.ts service
    - Base URL: https://nominatim.openstreetmap.org
    - Endpoint: /search?format=json&q={query}&limit=5
    - Add User-Agent header (required by Nominatim)
    - Define NominatimResult interface (display_name, lat, lon)
  - [x] 5.3 Create useGeocoding hook
    - Accept query string parameter
    - Implement 300ms debounce using useDebounce
    - Use @tanstack/react-query for caching
    - Return { suggestions, isLoading, error }
  - [x] 5.4 Create useDebounce utility hook
    - Generic debounce hook for any value
    - Configurable delay (default 300ms)
    - Path: /frontend/src/hooks/useDebounce.ts
  - [x] 5.5 Ensure geocoding tests pass
    - Run ONLY the 3-4 tests written in 5.1

**Acceptance Criteria:**
- Nominatim service fetches address suggestions
- useGeocoding hook debounces requests (300ms)
- Suggestions include display_name, lat, lon
- Error states handled gracefully
- Tests from 5.1 pass

---

#### Task Group 6: Location Picker Map Component
**Dependencies:** Task Group 5
**Complexity:** Medium-High
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/frontend/src/components/map/LocationPickerMap.tsx` (new)

- [x] 6.0 Complete interactive map component for location selection
  - [x] 6.1 Write 3-4 focused tests for LocationPickerMap
    - Test map renders with initial center from GOIAS_BOUNDS
    - Test click on map updates coordinates
    - Test marker is draggable and updates coordinates on dragend
    - Test onLocationChange callback fires with new coordinates
  - [x] 6.2 Create LocationPickerMap component structure
    - Props: initialPosition?: { lat, lng }, onLocationChange: (lat, lng) => void
    - Use dynamic import to avoid SSR issues
    - Follow pattern from MapContainer.tsx
  - [x] 6.3 Implement clean map without other hospitals
    - Single TileLayer from OpenStreetMap
    - No hospital markers from useMapHospitals
    - Compact size suitable for drawer (height: 200-250px)
  - [x] 6.4 Implement click-to-place marker
    - useMapEvents hook to capture click event
    - Place/move marker to clicked position
    - Call onLocationChange with new coordinates
  - [x] 6.5 Implement draggable marker
    - Marker with draggable={true}
    - eventHandlers={{ dragend: handler }}
    - Update coordinates on drag end
    - Use Leaflet default marker icon
  - [x] 6.6 Implement center map on position change
    - When initialPosition prop changes, center map
    - Animate fly-to for smooth UX
  - [x] 6.7 Ensure LocationPickerMap tests pass
    - Run ONLY the 3-4 tests written in 6.1

**Acceptance Criteria:**
- Map renders without SSR errors
- Click on map places/moves marker
- Marker is draggable with coordinate updates
- onLocationChange callback provides accurate lat/lng
- Map centers when initialPosition changes
- Tests from 6.1 pass

---

#### Task Group 7: Hospital Form Component
**Dependencies:** Task Groups 4, 5, 6
**Complexity:** High
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/frontend/src/components/forms/HospitalForm.tsx` (new)

- [x] 7.0 Complete hospital registration/edit form
  - [x] 7.1 Write 4-5 focused tests for HospitalForm
    - Test form renders all required fields
    - Test form validation prevents submit without coordinates
    - Test address autocomplete updates coordinates
    - Test submit calls onSubmit with form data
    - Test edit mode pre-fills form with hospital data
  - [x] 7.2 Create form schema with zod
    - nome: min 3 chars, required
    - codigo: required, alphanumeric pattern
    - endereco: required
    - telefone: optional, Brazilian phone regex
    - latitude: required, range -90 to 90
    - longitude: required, range -180 to 180
    - ativo: boolean, default true
  - [x] 7.3 Create HospitalForm component
    - Props: hospital?: Hospital (for edit), onSubmit, onCancel, isLoading
    - Use react-hook-form with zodResolver
    - Follow pattern from LoginForm.tsx
  - [x] 7.4 Implement form fields
    - Nome: Input with FormField wrapper
    - Codigo: Input with FormField wrapper
    - Endereco: Combobox/Input with autocomplete dropdown
    - Telefone: Input with phone mask (react-input-mask or manual)
    - Latitude: HIDDEN - auto-populated via map/autocomplete
    - Longitude: HIDDEN - auto-populated via map/autocomplete
    - Ativo: Switch component
  - [x] 7.5 Implement address autocomplete UI
    - Input field triggers useGeocoding on change
    - Dropdown shows suggestions below input
    - Click suggestion: fill endereco, set coordinates, center map
    - Use Popover or simple dropdown for suggestions
  - [x] 7.6 Integrate LocationPickerMap
    - Embed map below address field
    - Sync map position with form coordinates
    - Map click/drag updates latitude/longitude fields
  - [x] 7.7 Implement bidirectional sync
    - Address selection -> update coordinates -> move map marker
    - Map click/drag -> update coordinate fields
    - Coordinates are HIDDEN (not visible to user)
  - [x] 7.8 Implement submit button state
    - Disable until coordinates are filled
    - Show loading spinner during submission
    - Use Loader2 icon pattern from LoginForm
  - [x] 7.9 Ensure HospitalForm tests pass
    - Run ONLY the 4-5 tests written in 7.1

**Acceptance Criteria:**
- All form fields render correctly
- Zod validation works for all fields
- Address autocomplete populates coordinates
- Map and form stay synchronized
- Submit disabled until coordinates filled
- Edit mode pre-fills all fields
- Latitude/longitude fields are hidden from user
- Tests from 7.1 pass

---

#### Task Group 8: Hospital Form Drawer Component
**Dependencies:** Task Group 7
**Complexity:** Medium
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/frontend/src/components/hospitals/HospitalFormDrawer.tsx` (new)

- [x] 8.0 Complete drawer wrapper for hospital form
  - [x] 8.1 Write 2-3 focused tests for HospitalFormDrawer
    - Test drawer opens when open={true}
    - Test drawer calls onClose when dismissed
    - Test drawer shows "Novo Hospital" title for create mode
    - Test drawer shows "Editar Hospital" title for edit mode
  - [x] 8.2 Create HospitalFormDrawer component
    - Props: open, onClose, hospital?: Hospital, onSuccess
    - Use Sheet component from shadcn/ui
    - Follow pattern from HospitalDrawer.tsx
  - [x] 8.3 Implement drawer structure
    - SheetContent side="right" className="w-full sm:max-w-md overflow-y-auto"
    - SheetHeader with dynamic title (Novo/Editar Hospital)
    - SheetDescription with contextual text
  - [x] 8.4 Integrate HospitalForm
    - Pass hospital prop for edit mode
    - Handle form submission (create or update)
    - Call onSuccess callback after successful mutation
    - Close drawer on success
  - [x] 8.5 Implement create/update logic
    - Detect mode based on hospital prop presence
    - Use createHospital mutation for new hospitals
    - Use updateHospital mutation for existing hospitals
    - Show toast on success/error (using sonner)
  - [x] 8.6 Ensure HospitalFormDrawer tests pass
    - Run ONLY the 2-3 tests written in 8.1

**Acceptance Criteria:**
- Drawer opens/closes correctly
- Title changes based on create/edit mode
- Form submission creates or updates hospital
- Success triggers onSuccess callback and closes drawer
- Error shows toast notification
- Tests from 8.1 pass

---

#### Task Group 9: Hospitals Page Integration
**Dependencies:** Task Group 8
**Complexity:** Medium
**Files to modify:**
- `/home/matheus_rubem/VitalConnect/frontend/src/app/dashboard/hospitals/page.tsx`

- [x] 9.0 Complete hospitals page with create/edit functionality
  - [x] 9.1 Write 3-4 focused tests for hospitals page
    - Test "Novo Hospital" button visible for Admin/Gestor
    - Test "Novo Hospital" button hidden for Operador
    - Test clicking "Novo Hospital" opens drawer
    - Test clicking hospital card opens edit drawer (Admin/Gestor only)
  - [x] 9.2 Add "Novo Hospital" button with permission check
    - Import useAuth hook
    - Check user.role is 'admin' or 'gestor'
    - Render Button only if permitted
    - Position: top-right of page header
    - Icon: Plus from lucide-react
  - [x] 9.3 Add HospitalFormDrawer state management
    - State: drawerOpen, selectedHospital
    - Open drawer for create: drawerOpen=true, selectedHospital=null
    - Open drawer for edit: drawerOpen=true, selectedHospital=hospital
  - [x] 9.4 Make hospital cards clickable for edit
    - Wrap Card in clickable div/button
    - Only enable click for Admin/Gestor roles
    - On click: setSelectedHospital and open drawer
    - Add hover state visual feedback
  - [x] 9.5 Add telefone display to hospital cards
    - Show telefone in CardContent if available
    - Use Phone icon from lucide-react
    - Format: (XX) XXXX-XXXX
  - [x] 9.6 Handle drawer success callback
    - On create/update success, drawer closes
    - useHospitals data refreshes automatically (cache invalidation)
    - Show success toast
  - [x] 9.7 Ensure hospitals page tests pass
    - Run ONLY the 3-4 tests written in 9.1

**Acceptance Criteria:**
- "Novo Hospital" button appears for Admin and Gestor only
- Button hidden for Operador role
- Clicking button opens create drawer
- Clicking hospital card opens edit drawer (with permissions)
- Hospital list refreshes after create/update
- Tests from 9.1 pass

---

### Testing

#### Task Group 10: Test Review and Gap Analysis
**Dependencies:** Task Groups 1-9
**Complexity:** Medium
**Files to modify:**
- Various test files as needed

- [x] 10.0 Review existing tests and fill critical gaps only
  - [x] 10.1 Review tests from Task Groups 1-9
    - Review 3-4 tests from database layer (Task 1.1)
    - Review 2-3 tests from repository layer (Task 2.1)
    - Review 3-4 tests from validation layer (Task 3.1)
    - Review 2-3 tests from frontend API (Task 4.1)
    - Review 3-4 tests from geocoding service (Task 5.1)
    - Review 3-4 tests from LocationPickerMap (Task 6.1)
    - Review 4-5 tests from HospitalForm (Task 7.1)
    - Review 2-3 tests from HospitalFormDrawer (Task 8.1)
    - Review 3-4 tests from hospitals page (Task 9.1)
    - Total existing tests: approximately 26-34 tests
  - [x] 10.2 Analyze test coverage gaps for THIS feature only
    - Identify critical user workflows that lack coverage
    - Focus on end-to-end flows: create hospital, edit hospital
    - Check permission-based access control coverage
    - Verify map-form synchronization scenarios
  - [x] 10.3 Write up to 10 additional strategic tests maximum
    - E2E test: Complete hospital creation flow
    - E2E test: Hospital edit with location adjustment
    - Integration test: Address search to map pin placement
    - Integration test: Map drag updates form coordinates
    - Permission test: Operador cannot see create button
    - Permission test: Operador cannot edit hospitals
    - Validation test: Form submit blocked without coordinates
    - API test: Backend rejects POST without coordinates
    - Error handling test: Nominatim failure gracefully handled
    - Cache test: Hospital list updates after mutation
  - [x] 10.4 Run feature-specific tests only
    - Run ONLY tests related to hospital registration feature
    - Expected total: approximately 36-44 tests maximum
    - Do NOT run the entire application test suite
    - Verify critical workflows pass

**Acceptance Criteria:**
- All feature-specific tests pass (approximately 36-44 tests total)
- Critical user workflows for this feature are covered
- No more than 10 additional tests added when filling in testing gaps
- Testing focused exclusively on hospital registration feature requirements

---

## Execution Order

Recommended implementation sequence:

```
Phase 1: Backend Foundation (Can run in parallel)
  Task Group 1: Database Migration and Model Updates
  Task Group 5: Geocoding Service (frontend, no backend dependency)

Phase 2: Backend Completion
  Task Group 2: Backend Repository Updates (depends on 1)
  Task Group 3: Backend Validation Enhancement (depends on 2)

Phase 3: Frontend Foundation
  Task Group 4: Frontend Types and API Service (depends on 3)
  Task Group 6: Location Picker Map Component (depends on 5)

Phase 4: Frontend Integration
  Task Group 7: Hospital Form Component (depends on 4, 5, 6)
  Task Group 8: Hospital Form Drawer Component (depends on 7)

Phase 5: Page Integration
  Task Group 9: Hospitals Page Integration (depends on 8)

Phase 6: Testing
  Task Group 10: Test Review and Gap Analysis (depends on 1-9)
```

## Complexity Summary

| Task Group | Complexity | Estimated Effort |
|------------|------------|------------------|
| 1. Database Migration | Low | 1-2 hours |
| 2. Repository Updates | Low | 1-2 hours |
| 3. Backend Validation | Low | 1-2 hours |
| 4. Frontend Types/API | Low | 1-2 hours |
| 5. Geocoding Service | Medium | 2-3 hours |
| 6. LocationPickerMap | Medium-High | 3-4 hours |
| 7. Hospital Form | High | 4-6 hours |
| 8. Form Drawer | Medium | 2-3 hours |
| 9. Page Integration | Medium | 2-3 hours |
| 10. Test Review | Medium | 2-3 hours |

**Total Estimated Effort:** 19-30 hours

## Key Technical Decisions

1. **Nominatim for Geocoding:** Free, no API key required, 300ms debounce to respect rate limits
2. **Dynamic Import for Map:** Avoid Next.js SSR hydration issues with Leaflet
3. **Bidirectional Sync:** Form coordinates and map marker stay synchronized
4. **Required Coordinates:** Backend enforces coordinates for POST, frontend disables submit until filled
5. **Phone Mask:** Brazilian format (XX) XXXX-XXXX or (XX) XXXXX-XXXX
6. **Role-based Access:** Admin and Gestor can create/edit, Operador view-only
7. **Cache Invalidation:** Mutations invalidate ['hospitals'] query for automatic refresh
8. **Hidden Coordinates:** Latitude/longitude are hidden fields, auto-populated via map or address autocomplete

## Files Created/Modified Summary

### New Files
- `/home/matheus_rubem/VitalConnect/backend/migrations/018_add_telefone_to_hospitals.sql`
- `/home/matheus_rubem/VitalConnect/frontend/src/services/nominatim.ts`
- `/home/matheus_rubem/VitalConnect/frontend/src/hooks/useGeocoding.ts`
- `/home/matheus_rubem/VitalConnect/frontend/src/hooks/useDebounce.ts`
- `/home/matheus_rubem/VitalConnect/frontend/src/components/map/LocationPickerMap.tsx`
- `/home/matheus_rubem/VitalConnect/frontend/src/components/forms/HospitalForm.tsx`
- `/home/matheus_rubem/VitalConnect/frontend/src/components/hospitals/HospitalFormDrawer.tsx`

### Modified Files
- `/home/matheus_rubem/VitalConnect/backend/internal/models/hospital.go`
- `/home/matheus_rubem/VitalConnect/backend/internal/repository/hospital_repository.go`
- `/home/matheus_rubem/VitalConnect/frontend/src/types/index.ts`
- `/home/matheus_rubem/VitalConnect/frontend/src/hooks/useHospitals.ts`
- `/home/matheus_rubem/VitalConnect/frontend/src/app/dashboard/hospitals/page.tsx`
