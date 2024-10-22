(defcolumns
  ST
  CT
  (BYTE :u8)
  (ARG))

;; In the first row, ST is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(vanish:first first ST)

;; ST either remains constant, or increments by one.
(vanish increment (*
                      ;; ST[k] == ST[k+1]
                      (- ST (shift ST 1))
                      ;; Or, ST[k]+1 == ST[k+1]
                      (- (+ 1 ST) (shift ST 1))))

;; Increment or reset counter
(vanish heartbeat
	;; Only When ST != 0
	(* ST
           ;; If CT[k] == 3
           (if (- 3 CT)
               ;; Then, CT[k+1] == 0
               (shift CT 1)
               ;; Else, CT[k]+1 == CT[k+1]
               (- (+ 1 CT) (shift CT 1)))))

;; Argument accumulates byte values.
(vanish accumulator
           ;; If CT[k] == 0
           (if CT
               ;; Then, ARG == BYTE
               (- ARG BYTE)
               ;; Else, ARG = BYTE[k] + 256*BYTE[k-1]
               (- ARG (+ BYTE (* 256 (shift ARG -1))))))
