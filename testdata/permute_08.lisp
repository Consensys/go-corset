(defcolumns
  (X :i16@prove)
  (Y :i16@prove))
(defpermutation (A B) ((+ X) (- Y)))
(defconstraint diag_ab () (* A (- (shift A 1) B)))
