(defpurefun (vanishes! e0) e0)
;;
(defpurefun (next X) (shift X 1))
(defpurefun (prev X) (shift X -1))
;;
(defpurefun (eq! x y) (- x y))
;;
(defpurefun (will-eq! e0 e1) (eq! (next e0) e1))
(defpurefun (will-inc! e0 offset) (will-eq! e0 (+ e0 offset)))
(defpurefun (will-remain-constant! e0) (will-eq! e0 e0))
;;
(defpurefun (if-eq-else x val then else) (if (eq! x val) then else))

;; ===================================================
;; Constraints
;; ===================================================
(defcolumns STAMP CT)

;; In the first row, STAMP is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(defconstraint first (:domain {0}) STAMP)

;; In the last row of a valid frame, the counter must have its max
;; value.  This ensures that all non-padding frames are complete.
(defconstraint last (:domain {-1} :guard STAMP) (eq! CT 3))

;; STAMP either remains constant, or increments by one.
(defconstraint increment () (*
                      ;; STAMP[k] == STAMP[k+1]
                      (will-inc! STAMP 0)
                      ;; Or, STAMP[k]+1 == STAMP[k+1]
                      (will-inc! STAMP 1)))

;; If STAMP changes, counter resets to zero.
(defconstraint reset () (*
                  ;; STAMP[k] == STAMP[k+1]
                  (will-remain-constant! STAMP)
                  ;; Or, CT[k+1] == 0
                  (vanishes! (next CT))))

;; If STAMP non-zero and reaches end-of-cycle, then stamp increments;
;; otherwise, counter increments.
(defconstraint heartbeat (:guard STAMP)
  ;; If CT == 3
  (if-eq-else CT 3
      ;; Then, STAMP[k]+1 == STAMP[k]
      (will-inc! STAMP 1)
      ;; Else, CT[k]+1 == CT[k]
      (will-inc! CT 1)))
