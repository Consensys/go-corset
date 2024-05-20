(column X)
(permute (Y) (+X))
;; Ensure sorted column increments by 1
(vanish increment (- (shift Y 1) (+ 1 Y)))
