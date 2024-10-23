(defcolumns X)
(module m1)
(defcolumns ST (X :u16))
(permute (Y) (+X))
;; Ensure sorted column increments by 1
(defconstraint increment () (* ST (- (shift Y 1) (+ 1 Y))))
