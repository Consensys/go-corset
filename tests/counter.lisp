(* (- STAMP (shift STAMP 1)) (- (+ 1 STAMP) (shift STAMP 1)))
(* (- STAMP (shift STAMP 1)) (shift CT 1))
(if (- 3 CT) (- (+ 1 STAMP) (shift STAMP 1)) (- (+ 1 CT) (shift CT 1)))
