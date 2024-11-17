(defcolumns STAMP CT)

;; In the first row, STAMP is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(defconstraint first (:domain {0}) STAMP)

;; In the last row of a valid frame, the counter must have its max
;; value.  This ensures that all non-padding frames are complete.
(defconstraint last (:domain {-1}) (* STAMP (- 3 CT)))

;; STAMP either remains constant, or increments by one.
(defconstraint increment () (*
                      ;; STAMP[k] == STAMP[k+1]
                      (- STAMP (shift STAMP 1))
                      ;; Or, STAMP[k]+1 == STAMP[k+1]
                      (- (+ 1 STAMP) (shift STAMP 1))))

;; If STAMP changes, counter resets to zero.
(defconstraint reset () (*
                  ;; STAMP[k] == STAMP[k+1]
                  (- STAMP (shift STAMP 1))
                  ;; Or, CT[k+1] == 0
                  (shift CT 1)))

;; If STAMP non-zero and reaches end-of-cycle, then stamp increments;
;; otherwise, counter increments.
(defconstraint heartbeat ()
  (if STAMP
      0
      ;; If CT == 3
      (if (- 3 CT)
          ;; Then, STAMP[k]+1 == STAMP[k]
          (- (+ 1 STAMP) (shift STAMP 1))
          ;; Else, CT[k]+1 == CT[k]
          (- (+ 1 CT) (shift CT 1)))))
