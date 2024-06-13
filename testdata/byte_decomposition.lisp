(column ST)
(column CT)
(column BYTE :u8)
(column ARG)

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
