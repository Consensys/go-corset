(defcolumns (X :i16))
(module m1)
(defcolumns (ST :i4) (X :i16@prove))
(defpermutation (Y) ((+ X)))
;; Ensure sorted column increments by 1
(defconstraint increment ()
  (== 0 (* ST (- (shift Y 1) (+ 1 Y)))))
