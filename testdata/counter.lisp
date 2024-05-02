(column STAMP)
(column CT)

;; In the first row, STAMP is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(vanish:first first STAMP)

;; STAMP either remains constant, or increments by one.
(vanish increment (*
                      ;; STAMP[k] == STAMP[k+1]
                      (- STAMP (shift STAMP 1))
                      ;; Or, STAMP[k]+1 == STAMP[k+1]
                      (- (+ 1 STAMP) (shift STAMP 1))))

;; If STAMP changes, counter resets to zero.
(vanish reset (*
                  ;; STAMP[k] == STAMP[k+1]
                  (- STAMP (shift STAMP 1))
                  ;; Or, CT[k+1] == 0
                  (shift CT 1)))

;; If STAMP non-zero and reaches end-of-cycle, then stamp increments;
;; otherwise, counter increments.
(vanish heartbeat
           (ifnot STAMP
                  ;; If CT == 3
                  (if (- 3 CT)
                      ;; Then, STAMP[k]+1 == STAMP[k]
                      (- (+ 1 STAMP) (shift STAMP 1))
                      ;; Else, CT[k]+1 == CT[k]
                      (- (+ 1 CT) (shift CT 1)))))
