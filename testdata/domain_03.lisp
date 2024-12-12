(defpurefun ((vanishes! :@loob) x) x)

(defcolumns STAMP)

;; STAMP[0] == 0
(defconstraint start (:domain {0}) (vanishes! STAMP))
;; STAMP[-1] == 5
(defconstraint end (:domain {-1}) (vanishes! (- STAMP 5)))
;; STAMP either remains constant, or increments by one.
(defconstraint increment ()
  (vanishes!
   (*
    ;; STAMP[k] == STAMP[k+1]
    (- STAMP (shift STAMP 1))
    ;; Or, STAMP[k]+1 == STAMP[k+1]
    (- (+ 1 STAMP) (shift STAMP 1)))))
