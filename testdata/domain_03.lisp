(defcolumns STAMP)

;; STAMP[0] == 0
(vanish:first start STAMP)
;; STAMP[-1] == 5
(vanish:last end (- STAMP 5))
;; STAMP either remains constant, or increments by one.
(vanish increment (*
                      ;; STAMP[k] == STAMP[k+1]
                      (- STAMP (shift STAMP 1))
                      ;; Or, STAMP[k]+1 == STAMP[k+1]
                      (- (+ 1 STAMP) (shift STAMP 1))))
