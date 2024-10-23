(defcolumns STAMP)

;; STAMP[0] == 0
(defconstraint start (:domain {0}) STAMP)
;; STAMP[-1] == 5
(defconstraint end (:domain {-1}) (- STAMP 5))
;; STAMP either remains constant, or increments by one.
(defconstraint increment () (*
                      ;; STAMP[k] == STAMP[k+1]
                      (- STAMP (shift STAMP 1))
                      ;; Or, STAMP[k]+1 == STAMP[k+1]
                      (- (+ 1 STAMP) (shift STAMP 1))))
