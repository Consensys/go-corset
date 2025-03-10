(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (X :i16))
(module m1)
(defcolumns (ST :i4) (X :i16@prove))
(defpermutation (Y) ((+ X)))
;; Ensure sorted column increments by 1
(defconstraint increment ()
  (vanishes! (* ST (- (shift Y 1) (+ 1 Y)))))
