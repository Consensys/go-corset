(defpurefun (vanishes! x) (== 0 x))
;;
(defcolumns
  (ST :i16@prove)
  (X :i16@prove)
  (Y :i16@prove))
(defpermutation (ST' A B) ((+ ST) (- X) (+ Y)))
(defconstraint diag_ab ()
  (vanishes! (* ST' (- (shift A 1) B))))
