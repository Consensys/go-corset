(defcolumns
  (ST :i32)
  (CT :i4)
  (BYTE :i8@prove)
  (ARG :i32))

;; In the first row, ST is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(defconstraint first (:domain {0}) (eq! ST 0))

;; In the last row of a valid frame, the counter must have its max
;; value.  This ensures that all non-padding frames are complete.
(defconstraint last (:domain {-1} :guard ST)
  ;; CT[$] == 3
  (eq! CT 3))

;; ST either remains constant, or increments by one.
(defconstraint increment ()
  (or!
   ;; ST[k] == ST[k+1]
   (eq! ST (shift ST 1))
   ;; Or, ST[k]+1 == ST[k+1]
   (eq! (+ 1 ST) (next ST))))

;; If ST changes, counter resets to zero.
(defconstraint reset ()
  (or!
   ;; ST[k] == ST[k+1]
   (eq! ST (shift ST 1))
   ;; Or, CT[k+1] == 0
   (eq! (next CT) 0)))

;; Increment or reset counter
(defconstraint heartbeat (:guard ST)
  ;; If CT[k] == 3
  (if-eq-else CT 3
              ;; Then, ST[k]+1 = ST[k+1]
              (eq! (next ST) (+ 1 ST))
              ;; Else, CT[k]+1 == CT[k+1]
              (eq! (+ 1 CT) (next CT))))

;; This should always be true for CT.
(defproperty ct-bound (or! (== ST 0) (<= CT 3)))

;; Argument accumulates byte values.
(defconstraint accumulator (:guard ST)
  ;; If CT[k] == 0
  (if-eq-else CT 0
              ;; Then, ARG == BYTE
              (eq! ARG BYTE)
              ;; Else, ARG = BYTE[k] + 256*BYTE[k-1]
              (eq! ARG (+ BYTE (* 256 (prev ARG))))))
