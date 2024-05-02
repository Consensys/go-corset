(column STAMP)
(column CT)

;; STAMP either remains constant, or increments by one.
(vanishing increment (*
                      ;; STAMP[k] == STAMP[k+1]
                      (- STAMP (shift STAMP 1))
                      ;; Or, STAMP[k]+1 == STAMP[k+1]
                      (- (+ 1 STAMP) (shift STAMP 1))))

;; If STAMP changes, counter resets to zero.
(vanishing reset (*
                  ;; STAMP[k] == STAMP[k+1]
                  (- STAMP (shift STAMP 1))
                  ;; Or, CT[k+1] == 0
                  (shift CT 1)))

;; If STAMP reaches end-of-cycle, then stamp increments; otherwise,
;; counter increments.
(vanishing heartbeat
           ;; If CT == 3
           (if (- 3 CT)
               ;; Then, STAMP[k]+1 == STAMP[k]
               (- (+ 1 STAMP) (shift STAMP 1))
               ;; Else, CT[k]+1 == CT[k]
               (- (+ 1 CT) (shift CT 1))))
