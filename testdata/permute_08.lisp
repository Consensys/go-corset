(defcolumns
  (X :i16@loob@prove)
  (Y :i16@loob@prove))
(defpermutation (A B) ((+ X) (- Y)))
(defconstraint diag_ab () (* A (- (shift A 1) B)))
