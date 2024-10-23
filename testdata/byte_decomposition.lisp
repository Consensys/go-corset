(defcolumns
  ST
  CT
  (BYTE :u8)
  (ARG))

;; In the first row, ST is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(defconstraint first (:domain {0}) ST)

;; In the last row of a valid frame, the counter must have its max
;; value.  This ensures that all non-padding frames are complete.
(defconstraint last (:domain {-1}) (* ST (- 3 CT)))

;; ST either remains constant, or increments by one.
(defconstraint increment () (*
                      ;; ST[k] == ST[k+1]
                      (- ST (shift ST 1))
                      ;; Or, ST[k]+1 == ST[k+1]
                      (- (+ 1 ST) (shift ST 1))))

;; If ST changes, counter resets to zero.
(defconstraint reset () (*
                  ;; ST[k] == ST[k+1]
                  (- ST (shift ST 1))
                  ;; Or, CT[k+1] == 0
                  (shift CT 1)))

;; Increment or reset counter
(defconstraint heartbeat ()
	;; Only When ST != 0
	(* ST
           ;; If CT[k] == 3
           (if (- 3 CT)
               ;; Then, CT[k+1] == 0
               (shift CT 1)
               ;; Else, CT[k]+1 == CT[k+1]
               (- (+ 1 CT) (shift CT 1)))))

;; Argument accumulates byte values.
(defconstraint accumulator ()
	;; Only When ST != 0
        (* ST
           ;; If CT[k] == 0
           (if CT
               ;; Then, ARG == BYTE
               (- ARG BYTE)
               ;; Else, ARG = BYTE[k] + 256*BYTE[k-1]
               (- ARG (+ BYTE (* 256 (shift ARG -1)))))))
