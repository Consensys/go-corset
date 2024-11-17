(defcolumns ST (X :i16@prove))
(defpermutation (Y) ((↓ X)))
;; Ensure sorted column increments by 1
(defconstraint increment () (* ST (- (shift Y 1) (+ 1 Y))))
